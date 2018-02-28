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

	"github.com/lawrencewoodman/ddataset"
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
)

type Experiment struct {
	Title          string
	File           fileinfo.FileInfo
	Train          *Mode
	Test           *Mode
	RuleGeneration rule.GenerationDescriber
	Aggregators    []aggregator.Spec
	Goals          []*goal.Goal
	SortOrder      []rhkassessment.SortOrder
	Category       string
	Tags           []string
	Rules          []rule.Rule
}

type descFile struct {
	Title          string             `yaml:"title"`
	Category       string             `yaml:"category"`
	Tags           []string           `yaml:"tags"`
	Train          *modeDesc          `yaml:"train"`
	Test           *modeDesc          `yaml:"test"`
	RuleGeneration ruleGenerationDesc `yaml:"ruleGeneration"`
	Aggregators    []*aggregator.Desc `yaml:"aggregators"`
	Goals          []string           `yaml:"goals"`
	SortOrder      []sortDesc         `yaml:"sortOrder"`
	Rules          []string           `yaml:"rules"`
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

// InvalidExtError indicates that a config file has an invalid extension
type InvalidExtError string

func (e InvalidExtError) Error() string {
	return "invalid extension: " + string(e)
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
	var train *Mode
	var test *Mode
	var err error

	if err := d.checkValid(); err != nil {
		return nil, err
	}

	allFields := []string{}

	if d.Train != nil {
		train, err = newMode("train", cfg, d.Train)
		if err != nil {
			return nil, fmt.Errorf("experiment field: train: %s", err)
		}
		allFields = append(allFields, d.Train.Dataset.Fields...)
	}
	if d.Test != nil {
		test, err = newMode("test", cfg, d.Test)
		if err != nil {
			return nil, fmt.Errorf("experiment field: test: %s", err)
		}
		allFields = append(allFields, d.Test.Dataset.Fields...)
	}

	goals, err := goal.MakeGoals(d.Goals)
	if err != nil {
		return nil, fmt.Errorf("experiment field: goals: %s", err)
	}
	aggregators, err := aggregator.MakeSpecs(allFields, d.Aggregators)
	if err != nil {
		return nil, fmt.Errorf("experiment field: aggregators: %s", err)
	}
	sortOrder, err := makeSortOrder(aggregators, d.SortOrder)
	if err != nil {
		return nil, fmt.Errorf("experiment field: sortOrder: %s", err)
	}
	rules, err := rule.MakeDynamicRules(d.Rules)
	if err != nil {
		return nil, fmt.Errorf("experiment field: rules: %s", err)
	}

	return &Experiment{
		Title: d.Title,
		File:  file,
		Train: train,
		Test:  test,
		RuleGeneration: ruleGeneration{
			fields:     d.RuleGeneration.Fields,
			arithmetic: d.RuleGeneration.Arithmetic,
		},
		Aggregators: aggregators,
		Goals:       goals,
		SortOrder:   sortOrder,
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
	modes := []*Mode{e.Train, e.Test}
	for _, m := range modes {
		if m != nil {
			if err := m.Dataset.Release(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Experiment) processTrainDataset(
	cfg *config.Config,
	pm *progress.Monitor,
) ([]rule.Rule, error) {
	reportProgress := func(msg string, percent float64) error {
		return pm.ReportProgress(e.File.Name(), report.Train, msg, percent)
	}
	noRules := []rule.Rule{}

	if err := reportProgress("Describing train dataset", 0); err != nil {
		return noRules, err
	}

	desc, err := description.DescribeDataset(e.Train.Dataset)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't describe train dataset: %s", err)
	}

	userRulesAss, err := e.assessRules(1, train, e.Rules, pm, cfg)
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

	ass, err := e.assessRules(2, train, generatedRules, pm, cfg)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}
	ass, err = ass.Merge(userRulesAss)
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

	newAss, err := e.assessRules(3, train, tweakableRules, pm, cfg)
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

	newAss, err = e.assessRules(4, train, reducedDPRules, pm, cfg)
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

	newAss, err = e.assessRules(5, train, combinedRules, pm, cfg)
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

	ass, err = ass.Merge(userRulesAss)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't merge assessments: %s", err)
	}

	ass.Sort(e.SortOrder)
	ass.Refine()
	ass = ass.TruncateRuleAssessments(10)

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
	pm *progress.Monitor,
	rules []rule.Rule,
) error {
	err :=
		pm.ReportProgress(e.File.Name(), report.Test, "Describing train dataset", 0)
	if err != nil {
		return err
	}
	desc, err := description.DescribeDataset(e.Test.Dataset)
	if err != nil {
		return fmt.Errorf("Couldn't describe test dataset: %s", err)
	}
	ass, err := e.assessRules(1, test, rules, pm, cfg)
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
	pm *progress.Monitor,
	l logger.Logger,
	ignoreWhen bool,
) error {
	rules := e.Rules

	reportProcessing := func(mode string) error {
		l.Info(
			fmt.Sprintf("Processing experiment: %s, mode: %s",
				e.File.Name(), mode),
		)
		err := pm.AddExperiment(e.File.Name(), e.Title, e.Tags, e.Category)
		if err != nil {
			return l.Error(err)
		}
		return nil
	}

	reportSuccess := func(mode string) error {
		l.Info(
			fmt.Sprintf("Successfully processed experiment: %s, mode: %s",
				e.File.Name(), mode),
		)
		if pmErr := pm.ReportSuccess(e.File.Name()); pmErr != nil {
			return l.Error(pmErr)
		}
		return nil
	}

	reportError := func(err error) error {
		pmErr := pm.AddExperiment(e.File.Name(), e.Title, e.Tags, e.Category)
		if pmErr != nil {
			return l.Error(pmErr)
		}
		logErr :=
			fmt.Errorf("Error processing experiment: %s, %s", e.File.Name(), err)
		l.Error(logErr)
		if pmErr := pm.ReportError(e.File.Name(), err); pmErr != nil {
			return l.Error(pmErr)
		}
		return nil
	}

	isFinished, stamp := pm.GetFinishStamp(e.File.Name())

	if e.Train != nil {
		ok, err := e.Train.ShouldProcess(e.File, isFinished, stamp)
		if err != nil {
			return reportError(err)
		}
		if ok || ignoreWhen {
			if err := reportProcessing("train"); err != nil {
				return err
			}
			trainRules, err := e.processTrainDataset(cfg, pm)
			if err != nil {
				return reportError(err)
			}
			rules = append(rules, trainRules...)
			if err := reportSuccess("train"); err != nil {
				return err
			}
		}
	}

	if e.Test != nil {
		ok, err := e.Test.ShouldProcess(e.File, isFinished, stamp)
		if err != nil {
			return reportError(err)
		}
		if ok || ignoreWhen {
			if err := reportProcessing("test"); err != nil {
				return err
			}
			if err := e.processTestDataset(cfg, pm, rules); err != nil {
				return reportError(err)
			}
			if err := reportSuccess("test"); err != nil {
				return err
			}
		}
	}

	return nil
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

func (e *descFile) checkValid() error {
	if len(e.Title) == 0 {
		return errors.New("experiment missing: title")
	}
	if e.Train == nil && e.Test == nil {
		return errors.New(
			"experiment missing either: train or test",
		)
	}
	if e.Train != nil {
		if e.Train.Dataset == nil {
			return errors.New(
				"experiment field: train: missing dataset",
			)
		}
	}
	if e.Test != nil {
		if e.Test.Dataset == nil {
			return errors.New(
				"experiment field: test: missing dataset",
			)
		}
	}
	return nil
}

func reportTrainAssessProgress(
	pm *progress.Monitor,
	filename string,
	stage int,
	recordNum int64,
	numRecords int64,
) error {
	progress := 100.0 * float64(recordNum) / float64(numRecords)
	msg := fmt.Sprintf("Assessing rules %d/%d", stage, assessRulesNumStages)
	return pm.ReportProgress(filename, report.Train, msg, progress)
}

func reportTestAssessProgress(
	pm *progress.Monitor,
	filename string,
	recordNum int64,
	numRecords int64,
) error {
	progress := 100.0 * float64(recordNum) / float64(numRecords)
	msg := fmt.Sprintf("Assessing rules")
	return pm.ReportProgress(filename, report.Test, msg, progress)
}

func (e *Experiment) assessRules(
	stage int,
	mode datasetKind,
	rules []rule.Rule,
	pm *progress.Monitor,
	cfg *config.Config,
) (*rhkassessment.Assessment, error) {
	var wg sync.WaitGroup
	var result *rhkassessment.Assessment
	var dataset ddataset.Dataset
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
	if mode == train {
		dataset = e.Train.Dataset
	} else {
		dataset = e.Test.Dataset
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
					pm,
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
					pm,
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
