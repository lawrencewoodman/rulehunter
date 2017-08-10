/*
	rulehunter - A server to find rules in data based on user specified goals
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>

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
	"github.com/lawrencewoodman/ddataset/dcache"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/lawrencewoodman/ddataset/dsql"
	"github.com/lawrencewoodman/ddataset/dtruncate"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rhkit/aggregator"
	rhkassessment "github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/goal"
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

type Experiment struct {
	Title          string
	Dataset        ddataset.Dataset
	RuleFields     []string
	RuleComplexity rule.Complexity
	Aggregators    []aggregator.Spec
	Goals          []*goal.Goal
	SortOrder      []rhkassessment.SortOrder
	When           *dexpr.Expr
	Tags           []string
	Rules          []rule.Rule
}

type descFile struct {
	Title          string             `yaml:"title"`
	Tags           []string           `yaml:"tags"`
	Dataset        string             `yaml:"dataset"`
	Csv            *csvDesc           `yaml:"csv"`
	Sql            *sqlDesc           `yaml:"sql"`
	Fields         []string           `yaml:"fields"`
	RuleFields     []string           `yaml:"ruleFields"`
	RuleComplexity ruleComplexity     `yaml:"ruleComplexity"`
	Aggregators    []*aggregator.Desc `yaml:"aggregators"`
	Goals          []string           `yaml:"goals"`
	SortOrder      []sortDesc         `yaml:"sortOrder"`
	// An expression that works out whether to run the experiment
	When  string   `yaml:"when"`
	Rules []string `yaml:"rules"`
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

type ruleComplexity struct {
	Arithmetic bool `yaml:"arithmetic"`
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

func New(cfg *config.Config, d *descFile) (*Experiment, error) {
	if err := d.checkValid(); err != nil {
		return nil, err
	}

	dataset, err := makeDataset(d)
	if err != nil {
		return nil, err
	}

	if cfg.MaxNumRecords >= 1 {
		dataset = dtruncate.New(dataset, cfg.MaxNumRecords)
	}

	if cfg.MaxNumCacheRecords >= 1 {
		dataset = dcache.New(dataset, cfg.MaxNumCacheRecords)
	}

	goals, err := goal.MakeGoals(d.Goals)
	if err != nil {
		return nil, fmt.Errorf("goals: %s", err)
	}
	aggregators, err := aggregator.MakeSpecs(dataset.Fields(), d.Aggregators)
	if err != nil {
		return nil, fmt.Errorf("aggregators: %s", err)
	}
	sortOrder, err := makeSortOrder(aggregators, d.SortOrder)
	if err != nil {
		return nil, fmt.Errorf("sortOrder: %s", err)
	}
	when, err := makeWhenExpr(d.When)
	if err != nil {
		return nil, InvalidWhenExprError(d.When)
	}
	rules, err := rule.MakeDynamicRules(d.Rules)
	if err != nil {
		return nil, fmt.Errorf("rules: %s", err)
	}

	return &Experiment{
		Title:          d.Title,
		Dataset:        dataset,
		RuleFields:     d.RuleFields,
		RuleComplexity: rule.Complexity{Arithmetic: d.RuleComplexity.Arithmetic},
		Aggregators:    aggregators,
		Goals:          goals,
		SortOrder:      sortOrder,
		When:           when,
		Tags:           d.Tags,
		Rules:          rules,
	}, nil
}

const assessRulesNumStages = 4

func Process(
	experimentFile fileinfo.FileInfo,
	cfg *config.Config,
	l logger.Logger,
	experimentProgress *progress.Experiment,
) error {
	reportExperimentFail := func(err error, errs ...error) error {
		l.Error(fmt.Sprintf("Failed processing experiment: %s - %s",
			experimentFile.Name(), err))
		switch len(errs) {
		case 0:
			return err
		case 1:
			return experimentProgress.ReportError(errs[0])
		}
		panic("wrong number of errors")
	}
	experimentFullFilename :=
		filepath.Join(cfg.ExperimentsDir, experimentFile.Name())
	experiment, err := load(cfg, experimentFullFilename)
	if err != nil {
		fullErr := fmt.Errorf("Can't load experiment: %s, %s",
			experimentFile.Name(), err)
		return experimentProgress.ReportError(fullErr)
	}
	ok, err := shouldProcess(experimentProgress, experimentFile, experiment.When)
	if err != nil || !ok {
		return err
	}

	l.Info(fmt.Sprintf("Processing experiment: %s", experimentFile.Name()))
	err = experimentProgress.UpdateDetails(experiment.Title, experiment.Tags)
	if err != nil {
		return reportExperimentFail(err)
	}

	err = experimentProgress.ReportProgress("Describing dataset", 0)
	if err != nil {
		return reportExperimentFail(err)
	}

	dDescription, err :=
		describeDataset(cfg, experimentFile.Name(), experiment.Dataset)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't describe dataset: %s", err)
		return reportExperimentFail(err, fullErr)
	}

	err = experimentProgress.ReportProgress("Generating rules", 0)
	if err != nil {
		return reportExperimentFail(err)
	}
	rules, err := rule.Generate(
		dDescription,
		experiment.RuleFields,
		experiment.RuleComplexity,
	)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't generate rules: %s", err)
		return reportExperimentFail(err, fullErr)
	}

	ass := rhkassessment.New()
	err = assessRules(
		1,
		ass,
		rules,
		experiment,
		experimentProgress,
		cfg,
	)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return reportExperimentFail(err, fullErr)
	}

	ass.Sort(experiment.SortOrder)
	ass.Refine()
	sortedRules := ass.Rules()

	err = experimentProgress.ReportProgress("Tweaking rules", 0)
	if err != nil {
		return reportExperimentFail(err)
	}
	tweakableRules := rule.Tweak(1, sortedRules, dDescription)

	err = assessRules(
		2,
		ass,
		tweakableRules,
		experiment,
		experimentProgress,
		cfg,
	)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return reportExperimentFail(err, fullErr)
	}

	ass.Sort(experiment.SortOrder)
	ass.Refine()

	sortedRules = ass.Rules()

	err = experimentProgress.ReportProgress("Reduce DP of rules", 0)
	if err != nil {
		return reportExperimentFail(err)
	}
	reducedDPRules := rule.ReduceDP(sortedRules)

	err = assessRules(
		3,
		ass,
		reducedDPRules,
		experiment,
		experimentProgress,
		cfg,
	)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return reportExperimentFail(err, fullErr)
	}

	ass.Sort(experiment.SortOrder)
	ass.Refine()

	numRulesToCombine := 50
	bestNonCombinedRules := ass.Rules(numRulesToCombine)
	combinedRules := rule.Combine(bestNonCombinedRules)

	err = assessRules(
		4,
		ass,
		combinedRules,
		experiment,
		experimentProgress,
		cfg,
	)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return reportExperimentFail(err, fullErr)
	}

	ass.Sort(experiment.SortOrder)
	ass.Refine()
	ass = ass.TruncateRuleAssessments(cfg.MaxNumReportRules)

	report := report.New(
		experiment.Title,
		ass,
		experiment.Aggregators,
		experiment.SortOrder,
		experimentFile.Name(),
		experiment.Tags,
	)
	if err := report.WriteJSON(cfg); err != nil {
		fullErr := fmt.Errorf("Couldn't write json report: %s", err)
		return reportExperimentFail(err, fullErr)
	}

	if err := experimentProgress.ReportSuccess(); err != nil {
		return reportExperimentFail(err)
	}

	l.Info("Successfully processed experiment: " + experimentFile.Name())
	return nil
}

func load(cfg *config.Config, filename string) (
	e *Experiment,
	err error,
) {
	var d *descFile

	ext := filepath.Ext(filename)
	switch ext {
	case ".json":
		d, err = loadJSON(filename)
	case ".yaml":
		d, err = loadYAML(filename)
	default:
		return nil, InvalidExtError(ext)
	}
	if err != nil {
		return nil, err
	}

	return New(cfg, d)
}

func loadJSON(filename string) (*descFile, error) {
	var e descFile
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

func loadYAML(filename string) (*descFile, error) {
	var e descFile
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func makeSortOrder(
	aggregators []aggregator.Spec,
	sortDescs []sortDesc,
) ([]rhkassessment.SortOrder, error) {
	r := make([]rhkassessment.SortOrder, len(sortDescs))
	for i, sod := range sortDescs {
		so, err :=
			rhkassessment.NewSortOrder(aggregators, sod.AggregatorName, sod.Direction)
		if err != nil {
			return []rhkassessment.SortOrder{}, err
		}
		r[i] = so
	}
	return r, nil
}

func makeDataset(e *descFile) (d ddataset.Dataset, err error) {
	switch e.Dataset {
	case "csv":
		d = dcsv.New(
			e.Csv.Filename,
			e.Csv.HasHeader,
			rune(e.Csv.Separator[0]),
			e.Fields,
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
		d = dsql.New(sqlHandler, e.Fields)
	default:
		return nil,
			fmt.Errorf("Experiment field: dataset, has invalid type: %s", e.Dataset)
	}
	return
}

func (e *descFile) checkValid() error {
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

func describeDataset(
	cfg *config.Config,
	filename string,
	dataset ddataset.Dataset,
) (*description.Description, error) {
	_description, err := description.DescribeDataset(dataset)
	if err != nil {
		return nil, err
	}

	fdFilename := filepath.Join(cfg.BuildDir, "descriptions", filename)
	if err := _description.WriteJSON(fdFilename); err != nil {
		return nil, err
	}
	return _description, nil
}

func assessRulesWorker(
	wg *sync.WaitGroup,
	ass *rhkassessment.Assessment,
	rules []rule.Rule,
	experiment *Experiment,
	jobs <-chan assessJob,
	results chan<- assessJobResult,
) {
	defer wg.Done()
	for j := range jobs {
		rulesPartial := rules[j.startRuleNum:j.endRuleNum]
		err := ass.AssessRules(
			experiment.Dataset,
			rulesPartial,
			experiment.Aggregators,
			experiment.Goals,
		)
		results <- assessJobResult{err: err}
	}
}

func assessCollectResults(
	experimentProgress *progress.Experiment,
	stage int,
	numJobs int,
	results <-chan assessJobResult,
) error {
	jobNum := 0
	for r := range results {
		jobNum++
		if r.err != nil {
			return r.err
		}
		err := reportProgress(experimentProgress, stage, jobNum, numJobs)
		if err != nil {
			return err
		}
	}
	return nil
}

func reportProgress(
	experimentProgress *progress.Experiment,
	stage int,
	jobNum int,
	numJobs int,
) error {
	progress := 100.0 * float64(jobNum) / float64(numJobs)
	msg := fmt.Sprintf("Assessing rules %d/%d", stage, assessRulesNumStages)
	return experimentProgress.ReportProgress(msg, progress)
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
	ass *rhkassessment.Assessment,
	rules []rule.Rule,
	experiment *Experiment,
	experimentProgress *progress.Experiment,
	cfg *config.Config,
) error {
	var wg sync.WaitGroup
	progressIntervals := 1000
	numRules := len(rules)
	jobs := make(chan assessJob, 100)
	results := make(chan assessJobResult, 100)

	if stage > assessRulesNumStages {
		panic("assessRules: stage > assessRulesNumStages")
	}

	if err := reportProgress(experimentProgress, stage, 0, 1); err != nil {
		return err
	}

	wg.Add(cfg.MaxNumProcesses)
	for i := 0; i < cfg.MaxNumProcesses; i++ {
		go assessRulesWorker(&wg, ass, rules, experiment, jobs, results)
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

	return assessCollectResults(
		experimentProgress,
		stage,
		numJobs,
		results,
	)
}

type assessJob struct {
	startRuleNum int
	endRuleNum   int
}

type assessJobResult struct {
	err error
}

func shouldProcess(
	experimentProgress *progress.Experiment,
	experimentFile fileinfo.FileInfo,
	whenExpr *dexpr.Expr,
) (bool, error) {
	isFinished, stamp := experimentProgress.GetFinishStamp()
	if isFinished && experimentFile.ModTime().After(stamp) {
		isFinished, stamp := false, time.Now()
		ok, err := evalWhenExpr(time.Now(), isFinished, stamp, whenExpr)
		return ok, err
	}
	ok, err := evalWhenExpr(time.Now(), isFinished, stamp, whenExpr)
	return ok, err
}
