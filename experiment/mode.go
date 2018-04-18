// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package experiment

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcopy"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/lawrencewoodman/ddataset/dsql"
	"github.com/lawrencewoodman/ddataset/dtruncate"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rhkit/aggregator"
	"github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/rule"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"github.com/vlifesystems/rulehunter/report"
)

type Mode interface {
	// Kind returns the report type of the mode
	Kind() report.ModeKind
	Dataset() ddataset.Dataset
	Release() error
	NumAssessRulesStages() int
}

type datasetDesc struct {
	CSV    *csvDesc `yaml:"csv"`
	SQL    *sqlDesc `yaml:"sql"`
	Fields []string `yaml:"fields"`
}

type csvDesc struct {
	Filename  string `yaml:"filename"`
	HasHeader bool   `yaml:"hasHeader"`
	Separator string `yaml:"separator"`
}

type sqlDesc struct {
	DriverName     string `yaml:"driverName"`
	DataSourceName string `yaml:"dataSourceName"`
	Query          string `yaml:"query"`
}

type InvalidWhenExprError string

func (e InvalidWhenExprError) Error() string {
	return "when field invalid: " + string(e)
}

func makeDataset(
	cfg *config.Config,
	dd *datasetDesc,
) (ddataset.Dataset, error) {
	// File mode permission:
	// No special permission bits
	// User: Read, Write Execute
	// Group: None
	// Other: None
	const modePerm = 0700
	var dataset ddataset.Dataset
	if dd.CSV != nil && dd.SQL != nil {
		return nil, errors.New("can't specify csv and sql source")
	}
	if dd.CSV == nil && dd.SQL == nil {
		return nil,
			errors.New("has no csv or sql field")
	}
	if dd.CSV != nil {
		if dd.CSV.Filename == "" {
			return nil, errors.New("csv: missing filename")
		}
		if dd.CSV.Separator == "" {
			return nil, errors.New("csv: missing separator")
		}
		dataset = dcsv.New(
			dd.CSV.Filename,
			dd.CSV.HasHeader,
			rune(dd.CSV.Separator[0]),
			dd.Fields,
		)
	} else if dd.SQL != nil {
		if dd.SQL.DriverName == "" {
			return nil, errors.New("sql: missing driverName")
		}
		if dd.SQL.DataSourceName == "" {
			return nil, errors.New("sql: missing dataSourceName")
		}
		if dd.SQL.Query == "" {
			return nil, errors.New("sql: missing query")
		}
		sqlHandler, err := newSQLHandler(
			dd.SQL.DriverName,
			dd.SQL.DataSourceName,
			dd.SQL.Query,
		)
		if err != nil {
			return nil, fmt.Errorf("sql: %s", err)
		}
		dataset = dsql.New(sqlHandler, dd.Fields)
	}

	if cfg.MaxNumRecords >= 1 {
		dataset = dtruncate.New(dataset, cfg.MaxNumRecords)
	}
	// Copy dataset to get stable version
	buildTmpDir := filepath.Join(cfg.BuildDir, "tmp")
	if err := os.MkdirAll(buildTmpDir, modePerm); err != nil {
		return nil, err
	}
	copyDataset, err := dcopy.New(dataset, buildTmpDir)
	if err != nil {
		return nil, err
	}
	return copyDataset, nil
}

func startWorkers(
	wg *sync.WaitGroup,
	cfg *config.Config,
	aggregators []aggregator.Spec,
	goals []*goal.Goal,
	rules []rule.Rule,
) (
	[]*assessment.Assessment,
	[]chan ddataset.Record,
	chan error,
) {
	numRules := len(rules)
	assessments := []*assessment.Assessment{}
	records := []chan ddataset.Record{}
	errors := make(chan error, cfg.MaxNumProcesses+1)
	ruleStep := numRules / cfg.MaxNumProcesses
	if ruleStep < cfg.MaxNumProcesses {
		ruleStep = cfg.MaxNumProcesses
	}
	for i := 0; i < numRules; i += ruleStep {
		nextI := i + ruleStep
		if nextI > numRules {
			nextI = numRules
		}
		a := assessment.New(aggregators, goals)
		a.AddRules(append(rules[i:nextI], rule.NewTrue()))
		assessments = append(assessments, a)
		recordC := make(chan ddataset.Record, 100)
		records = append(records, recordC)

		// wg.Add here because sometimes wg.Wait() called before all
		// the goroutines had started
		wg.Add(1)
		go assessRulesWorker(wg, a, recordC, errors)
	}
	return assessments, records, errors
}

func sendRecordsToWorkers(
	wg *sync.WaitGroup,
	q *quitter.Quitter,
	reportProgress func(int64, int64) error,
	records []chan ddataset.Record,
	errors chan error,
	dataset ddataset.Dataset,
) error {
	const progressIntervals = int64(100)
	reportNumRecords := dataset.NumRecords() / progressIntervals
	if reportNumRecords == 0 {
		reportNumRecords = 1
	}
	conn, err := dataset.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	recordNum := int64(0)
	for conn.Next() {
		select {
		case <-q.C:
			return ErrQuitReceived
		case err := <-errors:
			return err
		default:
			break
		}
		record := conn.Read().Clone()
		for _, r := range records {
			r <- record
		}
		recordNum++
		if recordNum == 0 || recordNum%reportNumRecords == 0 {
			if err := reportProgress(recordNum, dataset.NumRecords()); err != nil {
				return err
			}
		}
	}
	return conn.Err()
}

func assessRules(
	e *Experiment,
	m Mode,
	stage int,
	rules []rule.Rule,
	pm *progress.Monitor,
	q *quitter.Quitter,
	cfg *config.Config,
) (*assessment.Assessment, error) {
	const subRulesStep = 1000
	var wg sync.WaitGroup
	var result *assessment.Assessment

	processSubRules := func(ruleProgress float64, subRules []rule.Rule) (
		*assessment.Assessment,
		error,
	) {
		prevMsg := ""
		prevProgress := 0.0
		reportProgress := func(recordNum, numRecords int64) error {
			msg :=
				fmt.Sprintf("Assessing rules %d/%d", stage, m.NumAssessRulesStages())
			progress :=
				100.0*ruleProgress - 1 + float64(recordNum)/float64(numRecords)
			if msg != prevMsg || progress-prevProgress >= 0.5 {
				prevMsg = msg
				prevProgress = progress
				return pm.ReportProgress(
					e.File.Name(),
					m.Kind(),
					msg,
					progress,
				)
			}
			return nil
		}

		assessments, records, errors :=
			startWorkers(&wg, cfg, e.Aggregators, e.Goals, subRules)
		err := sendRecordsToWorkers(
			&wg,
			q,
			reportProgress,
			records,
			errors,
			m.Dataset(),
		)

		// We have finished with records and errors now, so it makes sense
		// to close these channels and wait for the goroutines to finish
		for _, r := range records {
			close(r)
		}
		wg.Wait()
		select {
		case errs := <-errors:
			return nil, errs
		default:
			close(errors)
			break
		}
		if err != nil {
			return nil, err
		}

		subResult := assessments[0]
		for _, a := range assessments[1:] {
			subResult, err = subResult.Merge(a)
			if err != nil {
				return nil, err
			}
		}
		subResult.Sort(e.SortOrder)
		subResult.Refine()
		return subResult, nil
	}

	if stage > m.NumAssessRulesStages() {
		panic("assessRules: stage > assessRulesNumStages")
	}

	if len(rules) == 0 {
		rules = []rule.Rule{rule.NewTrue()}
	}

	for i := 0; i < len(rules); i += subRulesStep {
		endI := i + subRulesStep
		if endI > len(rules) {
			endI = len(rules)
		}
		ruleProgress := float64(endI) / float64(len(rules))
		subRules := rules[i:endI]
		subRules = append(subRules, rule.NewTrue())
		newAss, err := processSubRules(ruleProgress, subRules)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			result = newAss
		} else {
			result, err = result.Merge(newAss)
			if err != nil {
				return nil, err
			}
		}
	}

	result.Sort(e.SortOrder)
	result.Refine()
	return result, nil
}

func assessRulesWorker(
	wg *sync.WaitGroup,
	ass *assessment.Assessment,
	records <-chan ddataset.Record,
	errors chan<- error,
) {
	defer wg.Done()

	for r := range records {
		if err := ass.ProcessRecord(r); err != nil {
			errors <- err
			return
		}
	}
	if err := ass.Update(); err != nil {
		errors <- err
	}
}

func shouldProcessMode(
	when *dexpr.Expr,
	file fileinfo.FileInfo,
	isFinished bool,
	pmStamp time.Time,
) (bool, error) {
	if isFinished && file.ModTime().After(pmStamp) {
		isFinished, pmStamp = false, time.Now()
	}
	return evalWhenExpr(time.Now(), isFinished, pmStamp, when)
}
