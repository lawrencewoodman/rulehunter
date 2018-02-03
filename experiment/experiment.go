// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package experiment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

func newExperiment(
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
		trainDataset, err = makeDataset("trainDataset", cfg, d.Fields, d.TrainDataset)
		if err != nil {
			return nil, err
		}
	}
	if d.TestDataset != nil {
		testDataset, err = makeDataset("testDataset", cfg, d.Fields, d.TestDataset)
		if err != nil {
			return nil, err
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

	return newExperiment(cfg, file, d)
}

func (e *Experiment) Release() error {
	datasets := []ddataset.Dataset{e.TrainDataset, e.TestDataset}
	for _, d := range datasets {
		if d != nil {
			if err := d.Release(); err != nil {
				return err
			}
		}
	}
	return nil
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

	ass, err := e.assessRules(1, train, e.Rules, pr, cfg)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}

	if err := reportProgress("Generating rules", 0); err != nil {
		return noRules, err
	}
	generatedRules, err := rule.Generate(desc, e.RuleGeneration)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't generate rules: %s", err)
	}

	newAss, err := e.assessRules(2, train, generatedRules, pr, cfg)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}
	ass, err = ass.Merge(newAss)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't merge assessments: %s", err)
	}

	ass.Sort(e.SortOrder)
	ass.Refine()
	sortedRules := ass.Rules()

	if err := reportProgress("Tweaking rules", 0); err != nil {
		return noRules, err
	}
	tweakableRules := rule.Tweak(1, sortedRules, desc)

	newAss, err = e.assessRules(3, train, tweakableRules, pr, cfg)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}
	ass, err = ass.Merge(newAss)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't merge assessments: %s", err)
	}

	ass.Sort(e.SortOrder)
	ass.Refine()
	sortedRules = ass.Rules()

	if err := reportProgress("Reduce DP of rules", 0); err != nil {
		return noRules, err
	}
	reducedDPRules := rule.ReduceDP(sortedRules)

	newAss, err = e.assessRules(4, train, reducedDPRules, pr, cfg)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}
	ass, err = ass.Merge(newAss)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't merge assessments: %s", err)
	}

	numRulesToCombine := 50
	ass.Sort(e.SortOrder)
	ass.Refine()
	bestNonCombinedRules := ass.Rules(numRulesToCombine)
	combinedRules := rule.Combine(bestNonCombinedRules)

	newAss, err = e.assessRules(5, train, combinedRules, pr, cfg)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}
	ass, err = ass.Merge(newAss)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't merge assessments: %s", err)
	}

	ass.Sort(e.SortOrder)
	ass.Refine()
	ass = ass.TruncateRuleAssessments(2)

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
	ass, err := e.assessRules(1, test, rules, pr, cfg)
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
	cfg *config.Config,
	fields []string,
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
		return nil, fmt.Errorf(
			"Experiment field: %s, can't specify csv and sql source",
			experimentField,
		)
	}
	if dd.CSV == nil && dd.SQL == nil {
		return nil,
			fmt.Errorf("Experiment field: %s, has no csv or sql field", experimentField)
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
		dataset = dcsv.New(
			dd.CSV.Filename,
			dd.CSV.HasHeader,
			rune(dd.CSV.Separator[0]),
			fields,
		)
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
		dataset = dsql.New(sqlHandler, fields)
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

func reportTrainAssessProgress(
	pr progressReporter,
	filename string,
	stage int,
	recordNum int64,
	numRecords int64,
) error {
	progress := 100.0 * float64(recordNum) / float64(numRecords)
	msg := fmt.Sprintf("Assessing rules %d/%d", stage, assessRulesNumStages)
	return pr.ReportProgress(filename, report.Train, msg, progress)
}

func reportTestAssessProgress(
	pr progressReporter,
	filename string,
	recordNum int64,
	numRecords int64,
) error {
	progress := 100.0 * float64(recordNum) / float64(numRecords)
	msg := fmt.Sprintf("Assessing rules")
	return pr.ReportProgress(filename, report.Test, msg, progress)
}

func (e *Experiment) assessRules(
	stage int,
	mode datasetKind,
	rules []rule.Rule,
	pr progressReporter,
	cfg *config.Config,
) (*rhkassessment.Assessment, error) {
	var wg sync.WaitGroup
	var result *rhkassessment.Assessment
	progressIntervals := int64(1000)
	numRules := len(rules)
	records := []chan ddataset.Record{}
	errors := make(chan error, cfg.MaxNumProcesses+1)
	defer close(errors)
	assessments := []*rhkassessment.Assessment{}
	closeRecords := func() {
		for _, r := range records {
			close(r)
		}
	}
	dataset := e.TrainDataset
	if mode == test {
		dataset = e.TestDataset
	}

	if stage > assessRulesNumStages {
		panic("assessRules: stage > assessRulesNumStages")
	}

	if len(rules) == 0 {
		a := rhkassessment.New(e.Aggregators, e.Goals)
		a.NumRecords = dataset.NumRecords()
		if err := a.Update(); err != nil {
			closeRecords()
			return nil, err
		}
		closeRecords()
		return a, nil
	}

	ruleStep := numRules / cfg.MaxNumProcesses
	if ruleStep < cfg.MaxNumProcesses {
		ruleStep = cfg.MaxNumProcesses
	}
	for i := 0; i < numRules; i += ruleStep {
		nextI := i + ruleStep
		if nextI > numRules {
			nextI = numRules
		}
		a := rhkassessment.New(e.Aggregators, e.Goals)
		a.AddRules(rules[i:nextI])
		assessments = append(assessments, a)
		wg.Add(1)
		recordC := make(chan ddataset.Record, 100)
		records = append(records, recordC)
		go assessRulesWorker(&wg, a, recordC, errors)
	}

	reportNumRecords := dataset.NumRecords() / progressIntervals
	if reportNumRecords == 0 {
		reportNumRecords = 1
	}
	conn, err := dataset.Open()
	if err != nil {
		closeRecords()
		wg.Wait()
		return nil, err
	}
	defer conn.Close()

	recordNum := int64(0)
	for conn.Next() {
		select {
		case err := <-errors:
			closeRecords()
			wg.Wait()
			return nil, err
		default:
			break
		}
		record := conn.Read().Clone()
		for _, r := range records {
			r <- record
		}
		recordNum++
		if recordNum == 0 || recordNum%reportNumRecords == 0 {
			if mode == train {
				err := reportTrainAssessProgress(
					pr,
					e.File.Name(),
					stage,
					recordNum,
					dataset.NumRecords(),
				)
				if err != nil {
					closeRecords()
					wg.Wait()
					return nil, err
				}
			} else {
				err := reportTestAssessProgress(
					pr,
					e.File.Name(),
					recordNum,
					dataset.NumRecords(),
				)
				if err != nil {
					closeRecords()
					wg.Wait()
					return nil, err
				}
			}
		}
	}
	if err := conn.Err(); err != nil {
		closeRecords()
		wg.Wait()
		return nil, err
	}

	closeRecords()
	wg.Wait()
	select {
	case err := <-errors:
		return nil, err
	default:
		break
	}

	result = assessments[0]
	for _, a := range assessments[1:] {
		result, err = result.Merge(a)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func assessRulesWorker(
	wg *sync.WaitGroup,
	ass *rhkassessment.Assessment,
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
