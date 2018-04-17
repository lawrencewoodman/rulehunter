// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package experiment

import (
	"fmt"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/rule"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"github.com/vlifesystems/rulehunter/report"
)

type TestMode struct {
	dataset ddataset.Dataset
	when    *dexpr.Expr
}

type testModeDesc struct {
	Dataset *datasetDesc `yaml:"dataset"`
	// An expression that works out whether to run the experiment for this mode
	When string `yaml:"when"`
}

func newTestMode(cfg *config.Config, desc *testModeDesc) (*TestMode, error) {
	d, err := makeDataset(cfg, desc.Dataset)
	if err != nil {
		return nil, fmt.Errorf("dataset: %s", err)
	}
	when, err := makeWhenExpr(desc.When)
	if err != nil {
		return nil, InvalidWhenExprError(desc.When)
	}
	return &TestMode{
		dataset: d,
		when:    when,
	}, nil
}

func (m *TestMode) Kind() report.ModeKind {
	return report.Test
}

func (m *TestMode) Release() error {
	if m == nil {
		return nil
	}
	return m.dataset.Release()
}

func (m *TestMode) Dataset() ddataset.Dataset {
	return m.dataset
}

func (m *TestMode) NumAssessRulesStages() int {
	return 1
}

func (m *TestMode) Process(
	e *Experiment,
	cfg *config.Config,
	pm *progress.Monitor,
	q *quitter.Quitter,
	rules []rule.Rule,
) error {
	err := pm.ReportProgress(
		e.File.Name(),
		report.Test,
		"Describing train dataset",
		0,
	)
	if err != nil {
		return err
	}
	desc, err := description.DescribeDataset(m.dataset)
	if err != nil {
		return fmt.Errorf("Couldn't describe test dataset: %s", err)
	}
	ass, err := assessRules(e, m, 1, rules, pm, q, cfg)
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
