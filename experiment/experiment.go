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
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/report"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type experimentFile struct {
	Title          string                         `yaml:"title"`
	Tags           []string                       `yaml:"tags"`
	Dataset        string                         `yaml:"dataset"`
	Csv            *csvDesc                       `yaml:"csv"`
	Sql            *sqlDesc                       `yaml:"sql"`
	FieldNames     []string                       `yaml:"fieldNames"`
	RuleFieldNames []string                       `yaml:"ruleFieldNames"`
	Aggregators    []*rhexperiment.AggregatorDesc `yaml:"aggregators"`
	Goals          []string                       `yaml:"goals"`
	SortOrder      []*sortDesc                    `yaml:"sortOrder"`
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
	experimentFile fileinfo.FileInfo,
	cfg *config.Config,
	l logger.Logger,
	progressMonitor *progress.ProgressMonitor,
) error {
	epr, err := progress.NewExperimentProgressReporter(
		progressMonitor,
		experimentFile.Name(),
	)
	if err != nil {
		return err
	}
	experimentFullFilename :=
		filepath.Join(cfg.ExperimentsDir, experimentFile.Name())
	experiment, tags, whenExpr, err :=
		loadExperiment(experimentFullFilename, cfg.MaxNumRecords)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't load experiment file: %s", err)
		return epr.ReportError(fullErr)
	}
	ok, err := shouldProcess(progressMonitor, experimentFile, whenExpr)
	if err != nil || !ok {
		return err
	}

	l.Info(fmt.Sprintf("Processing experiment: %s", experimentFile.Name()))
	err = epr.UpdateDetails(experiment.Title, tags)
	if err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		return err
	}

	if err := epr.ReportProgress("Describing dataset", 0); err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		return err
	}
	fieldDescriptions, err := rhkit.DescribeDataset(experiment.Dataset)
	if err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		fullErr := fmt.Errorf("Couldn't describe dataset: %s", err)
		return epr.ReportError(fullErr)
	}

	if err := epr.ReportProgress("Generating rules", 0); err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		return err
	}
	rules := rhkit.GenerateRules(fieldDescriptions, experiment.RuleFieldNames)

	assessment, err := assessRules(1, rules, experiment, epr, cfg)
	if err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment.Sort(experiment.SortOrder)
	assessment.Refine()
	sortedRules := assessment.GetRules()

	if err := epr.ReportProgress("Tweaking rules", 0); err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		return err
	}
	tweakableRules := rhkit.TweakRules(1, sortedRules, fieldDescriptions)

	assessment2, err := assessRules(2, tweakableRules, experiment, epr, cfg)
	if err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment3, err := assessment.Merge(assessment2)
	if err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		fullErr := fmt.Errorf("Couldn't merge assessments: %s", err)
		return epr.ReportError(fullErr)
	}
	assessment3.Sort(experiment.SortOrder)
	assessment3.Refine()

	numRulesToCombine := 50
	bestNonCombinedRules := assessment3.GetRules(numRulesToCombine)
	combinedRules := rhkit.CombineRules(bestNonCombinedRules)

	assessment4, err := assessRules(3, combinedRules, experiment, epr, cfg)
	if err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment5, err := assessment3.Merge(assessment4)
	if err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		fullErr := fmt.Errorf("Couldn't merge assessments: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment5.Sort(experiment.SortOrder)
	assessment5.Refine()
	assessment6 := assessment5.TruncateRuleAssessments(cfg.MaxNumReportRules)

	err = report.WriteJson(
		assessment6,
		experiment,
		experimentFile.Name(),
		tags,
		cfg,
	)
	if err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		fullErr := fmt.Errorf("Couldn't write json report: %s", err)
		return epr.ReportError(fullErr)
	}

	if err := epr.ReportSuccess(); err != nil {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		return err
	}

	l.Info(
		fmt.Sprintf("Successfully processed experiment: %s", experimentFile.Name()),
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
		Title:       e.Title,
		Dataset:     dataset,
		RuleFields:  e.RuleFieldNames,
		Aggregators: e.Aggregators,
		Goals:       e.Goals,
		SortOrder:   makeRHSortOrder(e.SortOrder),
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

func assessRulesWorker(
	wg *sync.WaitGroup,
	rules []rule.Rule,
	experiment *rhexperiment.Experiment,
	jobs <-chan assessJob,
	results chan<- assessJobResult,
) {
	defer wg.Done()
	for j := range jobs {
		rulesPartial := rules[j.startRuleNum:j.endRuleNum]
		assessment, err := rhkit.AssessRules(rulesPartial, experiment)
		if err != nil {
			results <- assessJobResult{assessment: nil, err: err}
			return
		}
		results <- assessJobResult{assessment: assessment, err: nil}
	}
}

func assessCollectResults(
	epr *progress.ExperimentProgressReporter,
	stage int,
	numJobs int,
	results <-chan assessJobResult,
) (*rhkit.Assessment, error) {
	var assessment *rhkit.Assessment
	var err error
	jobNum := 0
	for r := range results {
		jobNum++
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
		if err := reportProgress(epr, stage, jobNum, numJobs); err != nil {
			return nil, err
		}
	}
	return assessment, nil
}

func reportProgress(
	epr *progress.ExperimentProgressReporter,
	stage int,
	jobNum int,
	numJobs int,
) error {
	progress := 100.0 * float64(jobNum) / float64(numJobs)
	msg := fmt.Sprintf("Assessing rules %d/%d", stage, assessRulesNumStages)
	return epr.ReportProgress(msg, progress)
}

func assessCreateJobs(numRules int, step int, jobs chan<- assessJob) {
	for i := 0; i < numRules; i += step {
		nextI := i + step
		if nextI > numRules {
			nextI = numRules
		}
		jobs <- assessJob{startRuleNum: i, endRuleNum: nextI}
	}
	close(jobs)
}

func assessRules(
	stage int,
	rules []rule.Rule,
	experiment *rhexperiment.Experiment,
	epr *progress.ExperimentProgressReporter,
	cfg *config.Config,
) (*rhkit.Assessment, error) {
	var wg sync.WaitGroup
	progressIntervals := 1000
	numRules := len(rules)
	jobs := make(chan assessJob, 100)
	results := make(chan assessJobResult, 100)

	if stage > assessRulesNumStages {
		panic("assessRules: stage > assessRulesNumStages")
	}

	if err := reportProgress(epr, stage, 0, 1); err != nil {
		return nil, err
	}

	wg.Add(cfg.MaxNumProcesses)
	for i := 0; i < cfg.MaxNumProcesses; i++ {
		go assessRulesWorker(&wg, rules, experiment, jobs, results)
	}

	if numRules < progressIntervals {
		progressIntervals = numRules
	}
	step := numRules / progressIntervals
	if step < 10 {
		step = 10
	}
	numJobs := int(math.Ceil(float64(numRules) / float64(step)))
	go assessCreateJobs(numRules, step, jobs)
	go func() {
		wg.Wait()
		close(results)
	}()

	assessment, err := assessCollectResults(
		epr,
		stage,
		numJobs,
		results,
	)
	return assessment, err
}

type assessJob struct {
	startRuleNum int
	endRuleNum   int
}

type assessJobResult struct {
	assessment *rhkit.Assessment
	err        error
}

func shouldProcess(
	pm *progress.ProgressMonitor,
	experimentFile fileinfo.FileInfo,
	whenExpr *dexpr.Expr,
) (bool, error) {
	isFinished, stamp := pm.GetFinishStamp(experimentFile.Name())
	if isFinished && experimentFile.ModTime().After(stamp) {
		isFinished, stamp := false, time.Now()
		ok, err := evalWhenExpr(time.Now(), isFinished, stamp, whenExpr)
		return ok, err
	}
	ok, err := evalWhenExpr(time.Now(), isFinished, stamp, whenExpr)
	return ok, err
}
