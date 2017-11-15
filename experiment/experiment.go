// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package experiment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	"github.com/vlifesystems/rulehunter/report"
	"gopkg.in/yaml.v2"
)

type Experiment struct {
	Title          string
	File           fileinfo.FileInfo
	TrainDataset   ddataset.Dataset
	TestDataset    ddataset.Dataset
	RuleGeneration rule.GenerationDescriber
	Aggregators    []aggregator.Spec
	Goals          []*goal.Goal
	SortOrder      []rhkassessment.SortOrder
	When           *dexpr.Expr
	Category       string
	Tags           []string
	Rules          []rule.Rule
}

type descFile struct {
	Title          string             `yaml:"title"`
	Category       string             `yaml:"category"`
	Tags           []string           `yaml:"tags"`
	TrainDataset   *datasetDesc       `yaml:"trainDataset"`
	TestDataset    *datasetDesc       `yaml:"testDataset"`
	Csv            *csvDesc           `yaml:"csv"`
	Sql            *sqlDesc           `yaml:"sql"`
	Fields         []string           `yaml:"fields"`
	RuleGeneration ruleGenerationDesc `yaml:"ruleGeneration"`
	Aggregators    []*aggregator.Desc `yaml:"aggregators"`
	Goals          []string           `yaml:"goals"`
	SortOrder      []sortDesc         `yaml:"sortOrder"`
	// An expression that works out whether to run the experiment
	When  string   `yaml:"when"`
	Rules []string `yaml:"rules"`
}

type datasetDesc struct {
	CSV *csvDesc `yaml:"csv"`
	SQL *sqlDesc `yaml:"sql"`
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

type ruleGenerationDesc struct {
	Fields     []string `yaml:"fields"`
	Arithmetic bool     `yaml:"arithmetic"`
}

type ruleGeneration struct {
	fields     []string
	arithmetic bool
}

func (rg ruleGeneration) Fields() []string {
	return rg.fields
}

func (rg ruleGeneration) Arithmetic() bool {
	return rg.arithmetic
}

type sortDesc struct {
	Aggregator string `yaml:"aggregator"`
	Direction  string `yaml:"direction"`
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

type progressReporter interface {
	ReportProgress(string, report.ModeKind, string, float64) error
}

// Which dataset is to be used
type datasetKind int

const (
	train datasetKind = iota
	test
)

func New(
	cfg *config.Config,
	file fileinfo.FileInfo,
	d *descFile,
) (*Experiment, error) {
	var trainDataset ddataset.Dataset
	var testDataset ddataset.Dataset
	var err error

	if err := d.checkValid(); err != nil {
		return nil, err
	}

	if d.TrainDataset != nil {
		trainDataset, err = makeDataset("trainDataset", d.Fields, d.TrainDataset)
		if err != nil {
			return nil, err
		}
	}
	if d.TestDataset != nil {
		testDataset, err = makeDataset("testDataset", d.Fields, d.TestDataset)
		if err != nil {
			return nil, err
		}
	}

	if cfg.MaxNumRecords >= 1 {
		if d.TrainDataset != nil {
			trainDataset = dtruncate.New(trainDataset, cfg.MaxNumRecords)
		}
		if d.TestDataset != nil {
			testDataset = dtruncate.New(testDataset, cfg.MaxNumRecords)
		}
	}

	if cfg.MaxNumCacheRecords >= 1 {
		if d.TrainDataset != nil {
			trainDataset = dcache.New(trainDataset, cfg.MaxNumCacheRecords)
		}
		if d.TestDataset != nil {
			testDataset = dcache.New(testDataset, cfg.MaxNumCacheRecords)
		}
	}

	goals, err := goal.MakeGoals(d.Goals)
	if err != nil {
		return nil, fmt.Errorf("goals: %s", err)
	}
	aggregators, err := aggregator.MakeSpecs(d.Fields, d.Aggregators)
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
		Title:        d.Title,
		File:         file,
		TrainDataset: trainDataset,
		TestDataset:  testDataset,
		RuleGeneration: ruleGeneration{
			fields:     d.RuleGeneration.Fields,
			arithmetic: d.RuleGeneration.Arithmetic,
		},
		Aggregators: aggregators,
		Goals:       goals,
		SortOrder:   sortOrder,
		When:        when,
		Tags:        d.Tags,
		Category:    d.Category,
		Rules:       rules,
	}, nil
}

const assessRulesNumStages = 5

func Load(cfg *config.Config, file fileinfo.FileInfo) (*Experiment, error) {
	var d *descFile
	var err error
	fullFilename := filepath.Join(cfg.ExperimentsDir, file.Name())

	ext := filepath.Ext(fullFilename)
	switch ext {
	case ".json":
		d, err = loadJSON(fullFilename)
	case ".yaml":
		d, err = loadYAML(fullFilename)
	default:
		return nil, InvalidExtError(ext)
	}
	if err != nil {
		return nil, err
	}

	return New(cfg, file, d)
}

func (e *Experiment) processTrainDataset(
	cfg *config.Config,
	pr progressReporter,
) ([]rule.Rule, error) {
	reportProgress := func(msg string, percent float64) error {
		return pr.ReportProgress(e.File.Name(), report.Train, msg, percent)
	}
	noRules := []rule.Rule{}

	if err := reportProgress("Describing train dataset", 0); err != nil {
		return noRules, err
	}

	desc, err := description.DescribeDataset(e.TrainDataset)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't describe train dataset: %s", err)
	}

	ass := rhkassessment.New()

	if err := e.assessRules(1, train, ass, e.Rules, pr, cfg); err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}

	if err := reportProgress("Generating rules", 0); err != nil {
		return noRules, err
	}
	generatedRules, err := rule.Generate(desc, e.RuleGeneration)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't generate rules: %s", err)
	}

	if err := e.assessRules(2, train, ass, generatedRules, pr, cfg); err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}

	ass.Sort(e.SortOrder)
	ass.Refine()
	sortedRules := ass.Rules()

	if err := reportProgress("Tweaking rules", 0); err != nil {
		return noRules, err
	}
	tweakableRules := rule.Tweak(1, sortedRules, desc)

	if err := e.assessRules(3, train, ass, tweakableRules, pr, cfg); err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}

	ass.Sort(e.SortOrder)
	ass.Refine()
	sortedRules = ass.Rules()

	if err := reportProgress("Reduce DP of rules", 0); err != nil {
		return noRules, err
	}
	reducedDPRules := rule.ReduceDP(sortedRules)

	if err := e.assessRules(4, train, ass, reducedDPRules, pr, cfg); err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}

	numRulesToCombine := 50
	ass.Sort(e.SortOrder)
	ass.Refine()
	bestNonCombinedRules := ass.Rules(numRulesToCombine)
	combinedRules := rule.Combine(bestNonCombinedRules)

	if err := e.assessRules(5, train, ass, combinedRules, pr, cfg); err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}

	ass.Sort(e.SortOrder)
	ass.Refine()
	ass = ass.TruncateRuleAssessments(cfg.MaxNumReportRules)

	r := report.New(
		report.Train,
		e.Title,
		desc,
		ass,
		e.Aggregators,
		e.SortOrder,
		e.File.Name(),
		e.Tags,
		e.Category,
	)
	if err := r.WriteJSON(cfg); err != nil {
		return noRules, fmt.Errorf("Couldn't write JSON train report: %s", err)
	}
	return ass.Rules(), nil
}

func (e *Experiment) processTestDataset(
	cfg *config.Config,
	pr progressReporter,
	rules []rule.Rule,
) error {
	err :=
		pr.ReportProgress(e.File.Name(), report.Test, "Describing train dataset", 0)
	if err != nil {
		return err
	}
	desc, err := description.DescribeDataset(e.TestDataset)
	if err != nil {
		return fmt.Errorf("Couldn't describe test dataset: %s", err)
	}
	ass := rhkassessment.New()
	err = e.assessRules(1, test, ass, rules, pr, cfg)
	if err != nil {
		return fmt.Errorf("Couldn't assess rules: %s", err)
	}
	testReport := report.New(
		report.Test,
		e.Title,
		desc,
		ass,
		e.Aggregators,
		e.SortOrder,
		e.File.Name(),
		e.Tags,
		e.Category,
	)
	if err := testReport.WriteJSON(cfg); err != nil {
		return fmt.Errorf("Couldn't write JSON test report: %s", err)
	}
	return nil
}

func (e *Experiment) Process(
	cfg *config.Config,
	pr progressReporter,
) error {
	rules := e.Rules
	if e.TrainDataset != nil {
		trainRules, err := e.processTrainDataset(cfg, pr)
		if err != nil {
			return err
		}
		rules = append(rules, trainRules...)
	}

	if e.TestDataset != nil {
		if err := e.processTestDataset(cfg, pr, rules); err != nil {
			return err
		}
	}

	return nil
}

func (e *Experiment) ShouldProcess(
	isFinished bool,
	stamp time.Time,
) (bool, error) {
	if isFinished && e.File.ModTime().After(stamp) {
		isFinished, stamp = false, time.Now()
	}
	return evalWhenExpr(time.Now(), isFinished, stamp, e.When)
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
			rhkassessment.NewSortOrder(aggregators, sod.Aggregator, sod.Direction)
		if err != nil {
			return []rhkassessment.SortOrder{}, err
		}
		r[i] = so
	}
	return r, nil
}

func makeDataset(
	experimentField string,
	fields []string,
	dd *datasetDesc,
) (ddataset.Dataset, error) {
	if dd.CSV != nil && dd.SQL != nil {
		return nil, fmt.Errorf(
			"Experiment field: %s, can't specify csv and sql source",
			experimentField,
		)
	}
	if dd.CSV != nil {
		if dd.CSV.Filename == "" {
			return nil, fmt.Errorf("Experiment field missing: %s > csv > filename",
				experimentField)
		}
		if dd.CSV.Separator == "" {
			return nil, fmt.Errorf("Experiment field missing: %s > csv > separator",
				experimentField)
		}
		return dcsv.New(
			dd.CSV.Filename,
			dd.CSV.HasHeader,
			rune(dd.CSV.Separator[0]),
			fields,
		), nil
	} else if dd.SQL != nil {
		if dd.SQL.DriverName == "" {
			return nil, fmt.Errorf(
				"Experiment field missing: %s > sql > driverName",
				experimentField,
			)
		}
		if dd.SQL.DataSourceName == "" {
			return nil, fmt.Errorf(
				"Experiment field missing: %s > sql > dataSourceName",
				experimentField,
			)
		}
		if dd.SQL.Query == "" {
			return nil, fmt.Errorf("Experiment field missing: %s > sql > query",
				experimentField)
		}
		sqlHandler, err := newSQLHandler(
			dd.SQL.DriverName,
			dd.SQL.DataSourceName,
			dd.SQL.Query,
		)
		if err != nil {
			return nil, fmt.Errorf("Experiment field: %s > sql, has %s",
				experimentField, err)
		}
		return dsql.New(sqlHandler, fields), nil
	}
	return nil,
		fmt.Errorf("Experiment field: %s, has no csv or sql field", experimentField)
}

func (e *descFile) checkValid() error {
	if len(e.Title) == 0 {
		return errors.New("Experiment field missing: title")
	}
	if e.TrainDataset == nil && e.TestDataset == nil {
		return errors.New(
			"Experiment field missing either: trainDataset or testDataset",
		)
	}
	return nil
}

func assessRulesWorker(
	mode datasetKind,
	wg *sync.WaitGroup,
	ass *rhkassessment.Assessment,
	rules []rule.Rule,
	e *Experiment,
	jobs <-chan assessJob,
	results chan<- assessJobResult,
) {
	defer wg.Done()
	dataset := e.TrainDataset
	if mode == test {
		dataset = e.TestDataset
	}
	for j := range jobs {
		rulesPartial := rules[j.startRuleNum:j.endRuleNum]
		err := ass.AssessRules(
			dataset,
			rulesPartial,
			e.Aggregators,
			e.Goals,
		)
		results <- assessJobResult{err: err}
	}
}

func assessCollectResults(
	pr progressReporter,
	filename string,
	mode datasetKind,
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
		if mode == train {
			err := reportTrainAssessProgress(pr, filename, stage, jobNum, numJobs)
			if err != nil {
				return err
			}
		} else {
			err := reportTestAssessProgress(pr, filename, jobNum, numJobs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func reportTrainAssessProgress(
	pr progressReporter,
	filename string,
	stage int,
	jobNum int,
	numJobs int,
) error {
	progress := 100.0 * float64(jobNum) / float64(numJobs)
	msg := fmt.Sprintf("Assessing rules %d/%d", stage, assessRulesNumStages)
	return pr.ReportProgress(filename, report.Train, msg, progress)
}

func reportTestAssessProgress(
	pr progressReporter,
	filename string,
	jobNum int,
	numJobs int,
) error {
	progress := 100.0 * float64(jobNum) / float64(numJobs)
	msg := fmt.Sprintf("Assessing rules")
	return pr.ReportProgress(filename, report.Test, msg, progress)
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

func (e *Experiment) assessRules(
	stage int,
	mode datasetKind,
	ass *rhkassessment.Assessment,
	rules []rule.Rule,
	pr progressReporter,
	cfg *config.Config,
) error {
	var wg sync.WaitGroup
	progressIntervals := 1000
	numRules := len(rules)
	jobs := make(chan assessJob, 100)
	results := make(chan assessJobResult, 100)

	if len(rules) == 0 {
		return nil
	}

	if stage > assessRulesNumStages {
		panic("assessRules: stage > assessRulesNumStages")
	}

	if mode == train {
		err := reportTrainAssessProgress(pr, e.File.Name(), stage, 0, 1)
		if err != nil {
			return err
		}
	} else {
		err := reportTestAssessProgress(pr, e.File.Name(), 0, 1)
		if err != nil {
			return err
		}
	}

	wg.Add(cfg.MaxNumProcesses)
	for i := 0; i < cfg.MaxNumProcesses; i++ {
		go assessRulesWorker(mode, &wg, ass, rules, e, jobs, results)
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

	return assessCollectResults(pr, e.File.Name(), mode, stage, numJobs, results)
}

type assessJob struct {
	startRuleNum int
	endRuleNum   int
}

type assessJobResult struct {
	err error
}
