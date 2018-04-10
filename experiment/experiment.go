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
	"github.com/vlifesystems/rulehunter/quitter"
	"github.com/vlifesystems/rulehunter/report"
	"gopkg.in/yaml.v2"
)

type Experiment struct {
	Title                string
	File                 fileinfo.FileInfo
	Train                *Mode
	Test                 *Mode
	RuleGeneration       ruleGeneration
	Aggregators          []aggregator.Spec
	Goals                []*goal.Goal
	SortOrder            []rhkassessment.SortOrder
	Category             string
	Tags                 []string
	Rules                []rule.Rule
	assessRulesNumStages int
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
	Fields            []string `yaml:"fields"`
	Arithmetic        bool     `yaml:"arithmetic"`
	CombinationLength int      `yaml:"combinationLength"`
}

type ruleGeneration struct {
	fields            []string
	arithmetic        bool
	combinationLength int
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

var ErrQuitReceived = errors.New("quit signal received")

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
			fields:            d.RuleGeneration.Fields,
			arithmetic:        d.RuleGeneration.Arithmetic,
			combinationLength: d.RuleGeneration.CombinationLength,
		},
		Aggregators:          aggregators,
		Goals:                goals,
		SortOrder:            sortOrder,
		Tags:                 d.Tags,
		Category:             d.Category,
		Rules:                rules,
		assessRulesNumStages: 4 + d.RuleGeneration.CombinationLength,
	}, nil
}

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
	q *quitter.Quitter,
) ([]rule.Rule, error) {
	reportProgress := func(msg string, percent float64) error {
		return pm.ReportProgress(e.File.Name(), report.Train, msg, percent)
	}
	quitReceived := func() bool {
		select {
		case <-q.C:
			return true
		default:
			return false
		}
	}
	noRules := []rule.Rule{}
	rt := newRuleTracker()

	if err := reportProgress("Describing train dataset", 0); err != nil {
		return noRules, err
	}

	if quitReceived() {
		return noRules, ErrQuitReceived
	}
	desc, err := description.DescribeDataset(e.Train.Dataset)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't describe train dataset: %s", err)
	}

	if quitReceived() {
		return noRules, ErrQuitReceived
	}
	rt.track(e.Rules)
	userRules := append(e.Rules, rule.NewTrue())
	ass, err := e.assessRules(1, train, userRules, pm, q, cfg)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't assess rules: %s", err)
	}

	assessRules := func(stage int, rules []rule.Rule) error {
		newRules := rt.track(rules)
		newAss, err := e.assessRules(stage, train, newRules, pm, q, cfg)
		if err != nil {
			return fmt.Errorf("Couldn't assess rules: %s", err)
		}
		ass, err = ass.Merge(newAss)
		if err != nil {
			return fmt.Errorf("Couldn't merge assessments: %s", err)
		}
		ass.Sort(e.SortOrder)
		ass.Refine()
		ass.TruncateRuleAssessments(5000)
		return nil
	}

	if err := reportProgress("Generating rules", 0); err != nil {
		return noRules, err
	}
	generatedRules, err := rule.Generate(desc, e.RuleGeneration)
	if err != nil {
		return noRules, fmt.Errorf("Couldn't generate rules: %s", err)
	}

	if quitReceived() {
		return noRules, ErrQuitReceived
	}

	if err := assessRules(2, generatedRules); err != nil {
		return noRules, err
	}

	if quitReceived() {
		return noRules, ErrQuitReceived
	}

	if err := reportProgress("Tweaking rules", 0); err != nil {
		return noRules, err
	}
	tweakableRules := rule.Tweak(1, ass.Rules(), desc)

	if err := assessRules(3, tweakableRules); err != nil {
		return noRules, err
	}

	if quitReceived() {
		return noRules, ErrQuitReceived
	}

	if err := reportProgress("Reduce DP of rules", 0); err != nil {
		return noRules, err
	}
	reducedDPRules := rule.ReduceDP(ass.Rules())

	if err := assessRules(4, reducedDPRules); err != nil {
		return noRules, err
	}

	if quitReceived() {
		return noRules, ErrQuitReceived
	}

	for i := 0; i < e.RuleGeneration.combinationLength; i++ {
		combinedRules := []rule.Rule{}
		hasCombined := false
		for n := 1000; !hasCombined || len(combinedRules) > 10000; n -= 10 {
			combinedRules = rule.Combine(ass.Rules(n))
			hasCombined = true
		}
		combinedRules = append(combinedRules, rule.NewTrue())

		if err := assessRules(5+i, combinedRules); err != nil {
			return noRules, err
		}

		if quitReceived() {
			return noRules, ErrQuitReceived
		}
	}

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
	pm *progress.Monitor,
	q *quitter.Quitter,
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
	ass, err := e.assessRules(1, test, rules, pm, q, cfg)
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
	q *quitter.Quitter,
	ignoreWhen bool,
) error {
	q.Add()
	defer q.Done()
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
			trainRules, err := e.processTrainDataset(cfg, pm, q)
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
			if err := e.processTestDataset(cfg, pm, q, rules); err != nil {
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

func (e *Experiment) startWorkers(
	wg *sync.WaitGroup,
	cfg *config.Config,
	rules []rule.Rule,
) (
	[]*rhkassessment.Assessment,
	[]chan ddataset.Record,
	chan error,
) {
	numRules := len(rules)
	assessments := []*rhkassessment.Assessment{}
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
		a := rhkassessment.New(e.Aggregators, e.Goals)
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

func (e *Experiment) sendRecordsToWorkers(
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

func (e *Experiment) assessRules(
	stage int,
	mode datasetKind,
	rules []rule.Rule,
	pm *progress.Monitor,
	q *quitter.Quitter,
	cfg *config.Config,
) (*rhkassessment.Assessment, error) {
	const subRulesStep = 1000
	var wg sync.WaitGroup
	var result *rhkassessment.Assessment
	var dataset ddataset.Dataset

	if mode == train {
		dataset = e.Train.Dataset
	} else {
		dataset = e.Test.Dataset
	}

	fmt.Printf("assessRules - stage: %d, len(rules): %d\n", stage, len(rules))

	processSubRules := func(ruleProgress float64, subRules []rule.Rule) (
		*rhkassessment.Assessment,
		error,
	) {
		reportProgress := func(recordNum, numRecords int64) error {
			progress :=
				100.0*ruleProgress - 1 + float64(recordNum)/float64(numRecords)
			if mode == train {
				msg :=
					fmt.Sprintf("Assessing rules %d/%d", stage, e.assessRulesNumStages)
				return pm.ReportProgress(e.File.Name(), report.Train, msg, progress)
			}
			msg := fmt.Sprintf("Assessing rules")
			return pm.ReportProgress(e.File.Name(), report.Test, msg, progress)
		}

		assessments, records, errors := e.startWorkers(&wg, cfg, rules)
		err :=
			e.sendRecordsToWorkers(&wg, q, reportProgress, records, errors, dataset)

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
		subResult.Sort(e.SortOrder)
		subResult.Refine()
		for _, a := range assessments[1:] {
			a.Sort(e.SortOrder)
			a.Refine()
			subResult, err = subResult.Merge(a)
			if err != nil {
				return nil, err
			}
		}

		return subResult, nil
	}

	if stage > e.assessRulesNumStages {
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
			newAss.Sort(e.SortOrder)
			newAss.Refine()
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
