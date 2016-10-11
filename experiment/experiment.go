/*
	rulehunter - A server to find rules in data based on user specified goals
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
	"github.com/vlifesystems/rhkit"
	rhexperiment "github.com/vlifesystems/rhkit/experiment"
	"github.com/vlifesystems/rhkit/rule"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/report"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type experimentFile struct {
	Title             string                         `yaml:"title"`
	Tags              []string                       `yaml:"tags"`
	Dataset           string                         `yaml:"dataset"`
	Csv               *csvDesc                       `yaml:"csv"`
	Sql               *sqlDesc                       `yaml:"sql"`
	FieldNames        []string                       `yaml:"fieldNames"`
	ExcludeFieldNames []string                       `yaml:"excludeFieldNames"`
	Aggregators       []*rhexperiment.AggregatorDesc `yaml:"aggregators"`
	Goals             []string                       `yaml:"goals"`
	SortOrder         []*sortDesc                    `yaml:"sortOrder"`
	// An expression that works out whether to run the experiment
	When string `yaml:"when"`
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

type sortDesc struct {
	AggregatorName string `yaml:"aggregatorName"`
	Direction      string `yaml:"direction"`
}

type InvalidWhenExprError string

func (e InvalidWhenExprError) Error() string {
	return "When field invalid: " + string(e)
}

// InvalidExtError indicates that a config file has an invalid extension
type InvalidExtError string

func (e InvalidExtError) Error() string {
	return "invalid extension: " + string(e)
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
	fieldDescriptions, err := rhkit.DescribeDataset(experiment.Dataset)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't describe dataset: %s", err)
		return epr.ReportError(fullErr)
	}

	if err := epr.ReportInfo("Generating rules"); err != nil {
		return err
	}
	rules, err :=
		rhkit.GenerateRules(fieldDescriptions, experiment.ExcludeFieldNames)
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
	tweakableRules := rhkit.TweakRules(sortedRules, fieldDescriptions)

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
	combinedRules := rhkit.CombineRules(bestNonCombinedRules)

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
	var e *experimentFile
	var noTags = []string{}

	ext := filepath.Ext(filename)
	switch ext {
	case ".json":
		e, err = loadJSON(filename)
	case ".yaml":
		e, err = loadYAML(filename)
	default:
		return nil, noTags, nil, InvalidExtError(ext)
	}
	if err != nil {
		return nil, noTags, nil, err
	}

	if err := e.checkValid(); err != nil {
		return nil, noTags, nil, err
	}

	dataset, err := makeDataset(e)
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
		SortOrder:     makeRHSortOrder(e.SortOrder),
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

func makeRHSortOrder(sortOrder []*sortDesc) []*rhexperiment.SortDesc {
	r := make([]*rhexperiment.SortDesc, len(sortOrder))
	for i, sd := range sortOrder {
		r[i] = &rhexperiment.SortDesc{
			AggregatorName: sd.AggregatorName,
			Direction:      sd.Direction,
		}
	}
	return r
}

func loadJSON(filename string) (*experimentFile, error) {
	var e experimentFile
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err = dec.Decode(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

func loadYAML(filename string) (*experimentFile, error) {
	var e experimentFile
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &e); err != nil {
		return nil, err
	}
	return &e, nil
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
		sqlHandler, err := newSQLHandler(
			e.Sql.DriverName,
			e.Sql.DataSourceName,
			e.Sql.Query,
		)
		if err != nil {
			return nil, fmt.Errorf("Experiment field: sql, has %s", err)
		}
		d = dsql.New(sqlHandler, e.FieldNames)
	default:
		return nil,
			fmt.Errorf("Experiment field: dataset, has invalid type: %s", e.Dataset)
	}
	return
}

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
) (*rhkit.Assessment, error) {
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
		assessment, err := rhkit.AssessRules(rules, experiment)
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
				assessment, err := rhkit.AssessRules(rulesPartial, experiment)
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

	var assessment *rhkit.Assessment
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
	assessment *rhkit.Assessment
	err        error
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
