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

	"github.com/vlifesystems/rhkit/aggregator"
	rhkassessment "github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/rule"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"gopkg.in/yaml.v2"
)

type Experiment struct {
	Title       string
	File        fileinfo.FileInfo
	Train       *TrainMode
	Test        *TestMode
	Aggregators []aggregator.Spec
	Goals       []*goal.Goal
	SortOrder   []rhkassessment.SortOrder
	Category    string
	Tags        []string
	Rules       []rule.Rule
}

type descFile struct {
	Title       string             `yaml:"title"`
	Category    string             `yaml:"category"`
	Tags        []string           `yaml:"tags"`
	Train       *trainModeDesc     `yaml:"train"`
	Test        *testModeDesc      `yaml:"test"`
	Aggregators []*aggregator.Desc `yaml:"aggregators"`
	Goals       []string           `yaml:"goals"`
	SortOrder   []sortDesc         `yaml:"sortOrder"`
	Rules       []string           `yaml:"rules"`
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
	var train *TrainMode
	var test *TestMode
	var err error

	if err := d.checkValid(); err != nil {
		return nil, err
	}

	allFields := []string{}

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

	if d.Train != nil {
		train, err = newTrainMode(
			cfg,
			d.Train,
			file.Name(),
			aggregators,
			goals,
			sortOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("experiment field: train: %s", err)
		}
		allFields = append(allFields, d.Train.Dataset.Fields...)
	}
	if d.Test != nil {
		test, err = newTestMode(cfg, d.Test)
		if err != nil {
			return nil, fmt.Errorf("experiment field: test: %s", err)
		}
		allFields = append(allFields, d.Test.Dataset.Fields...)
	}

	rules, err := rule.MakeDynamicRules(d.Rules)
	if err != nil {
		return nil, fmt.Errorf("experiment field: rules: %s", err)
	}

	return &Experiment{
		Title:       d.Title,
		File:        file,
		Train:       train,
		Test:        test,
		Aggregators: aggregators,
		Goals:       goals,
		SortOrder:   sortOrder,
		Tags:        d.Tags,
		Category:    d.Category,
		Rules:       rules,
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
	modes := []Mode{e.Train, e.Test}
	for _, m := range modes {
		if m != nil {
			if err := m.Release(); err != nil {
				return err
			}
		}
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
		ok, err := shouldProcessMode(e.Train.when, e.File, isFinished, stamp)
		if err != nil {
			return reportError(err)
		}
		if ok || ignoreWhen {
			if err := reportProcessing("train"); err != nil {
				return err
			}
			trainRules, err := e.Train.Process(e, cfg, pm, q, rules)
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
		ok, err := shouldProcessMode(e.Test.when, e.File, isFinished, stamp)
		if err != nil {
			return reportError(err)
		}
		if ok || ignoreWhen {
			if err := reportProcessing("test"); err != nil {
				return err
			}
			if err := e.Test.Process(e, cfg, pm, q, rules); err != nil {
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
