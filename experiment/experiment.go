/*
	rulehuntersrv - A server to find rules in data based on user specified goals
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

package experiment

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/lawrencewoodman/ddataset/dsql"
	"github.com/lawrencewoodman/ddataset/dtruncate"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rulehunter"
	rhexperiment "github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehunter/rule"
	"github.com/vlifesystems/rulehuntersrv/config"
	"github.com/vlifesystems/rulehuntersrv/logger"
	"github.com/vlifesystems/rulehuntersrv/progress"
	"github.com/vlifesystems/rulehuntersrv/report"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type experimentFile struct {
	Title             string
	Tags              []string
	Dataset           string
	Csv               *csvDesc
	Sql               *sqlDesc
	FieldNames        []string
	ExcludeFieldNames []string
	Aggregators       []*rhexperiment.AggregatorDesc
	Goals             []string
	SortOrder         []*rhexperiment.SortDesc
	// An expression that works out whether to run the experiment
	When string
}

type csvDesc struct {
	Filename  string
	HasHeader bool
	Separator string
}

type sqlDesc struct {
	DriverName     string
	DataSourceName string
	Query          string
}

type InvalidWhenExprError string

func (e InvalidWhenExprError) Error() string {
	return "When field invalid: " + string(e)
}

const assessRulesNumStages = 3

func Process(
	experimentFilename string,
	cfg *config.Config,
	l logger.Logger,
	progressMonitor *progress.ProgressMonitor,
) error {
	epr, err := progress.NewExperimentProgressReporter(
		progressMonitor,
		experimentFilename,
	)
	if err != nil {
		return err
	}
	experimentFullFilename :=
		filepath.Join(cfg.ExperimentsDir, experimentFilename)
	experiment, tags, whenExpr, err :=
		loadExperiment(experimentFullFilename, cfg.MaxNumRecords)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't load experiment file: %s", err)
		return epr.ReportError(fullErr)
	}
	ok, err := shouldProcess(progressMonitor, experimentFilename, whenExpr)
	if err != nil || !ok {
		return err
	}

	l.Info(fmt.Sprintf("Processing experiment: %s", experimentFilename))

	err = epr.UpdateDetails(experiment.Title, tags)
	if err != nil {
		return err
	}

	if err := epr.ReportInfo("Describing dataset"); err != nil {
		return err
	}
	fieldDescriptions, err := rulehunter.DescribeDataset(experiment.Dataset)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't describe dataset: %s", err)
		return epr.ReportError(fullErr)
	}

	if err := epr.ReportInfo("Generating rules"); err != nil {
		return err
	}
	rules, err :=
		rulehunter.GenerateRules(fieldDescriptions, experiment.ExcludeFieldNames)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't generate rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment, err := assessRules(1, rules, experiment, epr, cfg)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment.Sort(experiment.SortOrder)
	assessment.Refine(3)
	sortedRules := assessment.GetRules()

	if err := epr.ReportInfo("Tweaking rules"); err != nil {
		return err
	}
	tweakableRules := rulehunter.TweakRules(sortedRules, fieldDescriptions)

	assessment2, err := assessRules(2, tweakableRules, experiment, epr, cfg)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment3, err := assessment.Merge(assessment2)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't merge assessments: %s", err)
		return epr.ReportError(fullErr)
	}
	assessment3.Sort(experiment.SortOrder)
	assessment3.Refine(1)

	numRulesToCombine := 50
	bestNonCombinedRules := assessment3.GetRules(numRulesToCombine)
	combinedRules := rulehunter.CombineRules(bestNonCombinedRules)

	assessment4, err := assessRules(3, combinedRules, experiment, epr, cfg)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment5, err := assessment3.Merge(assessment4)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't merge assessments: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment5.Sort(experiment.SortOrder)
	assessment5.Refine(1)
	assessment6 := assessment5.TruncateRuleAssessments(cfg.MaxNumReportRules)

	err = report.WriteJson(
		assessment6,
		experiment,
		experimentFilename,
		tags,
		cfg,
	)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't write json report: %s", err)
		return epr.ReportError(fullErr)
	}

	if err := epr.ReportSuccess(); err != nil {
		return err
	}

	l.Info(
		fmt.Sprintf("Successfully processed experiment: %s", experimentFilename),
	)
	return nil
}

func loadExperiment(filename string, maxNumRecords int) (
	experiment *rhexperiment.Experiment,
	tags []string,
	whenExpr *dexpr.Expr,
	err error,
) {
	var f *os.File
	var e experimentFile
	var noTags = []string{}

	f, err = os.Open(filename)
	if err != nil {
		return nil, noTags, nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err = dec.Decode(&e); err != nil {
		return nil, noTags, nil, err
	}

	if err := e.checkValid(); err != nil {
		return nil, noTags, nil, err
	}

	dataset, err := makeDataset(&e)
	if err != nil {
		return nil, noTags, nil, err
	}

	if maxNumRecords >= 1 {
		dataset = dtruncate.New(dataset, maxNumRecords)
	}

	experimentDesc := &rhexperiment.ExperimentDesc{
		Title:         e.Title,
		Dataset:       dataset,
		ExcludeFields: e.ExcludeFieldNames,
		Aggregators:   e.Aggregators,
		Goals:         e.Goals,
		SortOrder:     e.SortOrder,
	}
	experiment, err = rhexperiment.New(experimentDesc)
	if err != nil {
		return nil, noTags, nil, err
	}

	whenExpr, err = makeWhenExpr(e.When)
	if err != nil {
		return nil, noTags, nil, InvalidWhenExprError(e.When)
	}
	return experiment, e.Tags, whenExpr, err
}

func makeDataset(e *experimentFile) (d ddataset.Dataset, err error) {
	switch e.Dataset {
	case "csv":
		d = dcsv.New(
			e.Csv.Filename,
			e.Csv.HasHeader,
			rune(e.Csv.Separator[0]),
			e.FieldNames,
		)
	case "sql":
		sqlHandler := newSQLHandler(
			e.Sql.DriverName,
			e.Sql.DataSourceName,
			e.Sql.Query,
		)
		d = dsql.New(sqlHandler, e.FieldNames)
	default:
		return nil,
			fmt.Errorf("Experiment field: dataset, has invalid type: %s", e.Dataset)
	}
	return
}

var validDrivers = []string{"sqlite3"}

func (e *experimentFile) checkValid() error {
	if len(e.Title) == 0 {
		return errors.New("Experiment field missing: title")
	}
	if len(e.Dataset) == 0 {
		return errors.New("Experiment field missing: dataset")
	}
	if e.Dataset == "csv" {
		if e.Csv == nil {
			return errors.New("Experiment field missing: csv")
		}
		if len(e.Csv.Filename) == 0 {
			return errors.New("Experiment field missing: csv > filename")
		}
		if len(e.Csv.Separator) == 0 {
			return errors.New("Experiment field missing: csv > separator")
		}
	}
	if e.Dataset == "sql" {
		if e.Sql == nil {
			return errors.New("Experiment field missing: sql")
		}
		if len(e.Sql.DriverName) == 0 {
			return errors.New("Experiment field missing: sql > driverName")
		}
		if !inStringsSlice(e.Sql.DriverName, validDrivers) {
			return errors.New("Experiment has invalid sql > driverName")
		}
		if len(e.Sql.DataSourceName) == 0 {
			return errors.New("Experiment field missing: sql > dataSourceName")
		}
		if len(e.Sql.Query) == 0 {
			return errors.New("Experiment field missing: sql > query")
		}
	}
	return nil
}

func assessRules(
	stage int,
	rules []rule.Rule,
	experiment *rhexperiment.Experiment,
	epr *progress.ExperimentProgressReporter,
	cfg *config.Config,
) (*rulehunter.Assessment, error) {
	var wg sync.WaitGroup
	var rulesProcessed uint64 = 0

	if stage > assessRulesNumStages {
		panic("assessRules: stage > assessRulesNumStages")
	}

	msg := fmt.Sprintf("Assessing rules %d/%d", stage, assessRulesNumStages)
	if err := epr.ReportInfo(msg); err != nil {
		return nil, err
	}

	numRules := len(rules)
	if numRules < 2 {
		assessment, err := rulehunter.AssessRules(rules, experiment)
		if err != nil {
			return nil, err
		}
		return assessment, err
	}

	progressIntervals := 1000
	if numRules < progressIntervals {
		progressIntervals = numRules
	}

	assessJobs := make(chan assessJob, progressIntervals)
	assessJobResults := make(chan assessJobOutcome, progressIntervals)

	for i := 0; i < cfg.MaxNumProcesses; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range assessJobs {
				rulesPartial := rules[j.startRuleNum:j.endRuleNum]
				assessment, err := rulehunter.AssessRules(rulesPartial, experiment)
				if err != nil {
					assessJobResults <- assessJobOutcome{assessment: nil, err: err}
					return
				}
				atomic.AddUint64(&rulesProcessed, uint64(len(assessment.RuleAssessments)))
				progress :=
					float64(atomic.LoadUint64(&rulesProcessed)) / float64(numRules) * 100.0
				msg := fmt.Sprintf("Assessing rules %d/%d: %.2f%%",
					stage, assessRulesNumStages, progress)
				if err := epr.ReportInfo(msg); err != nil {
					assessJobResults <- assessJobOutcome{assessment: nil, err: err}
					return
				}
				assessJobResults <- assessJobOutcome{assessment: assessment, err: nil}
			}
		}()
	}

	step := numRules / progressIntervals
	for i := 0; i < numRules; i += step {
		nextI := i + step
		if nextI > numRules {
			nextI = numRules
		}
		assessJobs <- assessJob{startRuleNum: i, endRuleNum: nextI}
	}
	close(assessJobs)

	go func() {
		wg.Wait()
		close(assessJobResults)
	}()

	var assessment *rulehunter.Assessment
	var err error
	for r := range assessJobResults {
		if r.err != nil {
			return nil, r.err
		}
		if assessment == nil {
			assessment = r.assessment
		} else {
			assessment, err = assessment.Merge(r.assessment)
			if err != nil {
				return nil, err
			}
		}
	}
	return assessment, nil
}

type assessJob struct {
	startRuleNum int
	endRuleNum   int
}

type assessJobOutcome struct {
	assessment *rulehunter.Assessment
	err        error
}

func inStringsSlice(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func shouldProcess(
	pm *progress.ProgressMonitor,
	experimentFilename string,
	whenExpr *dexpr.Expr,
) (bool, error) {
	isFinished, stamp := pm.GetFinishStamp(experimentFilename)
	ok, err := evalWhenExpr(time.Now(), isFinished, stamp, whenExpr)
	return ok, err
}
