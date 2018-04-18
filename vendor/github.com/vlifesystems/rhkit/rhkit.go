// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// package rhkit is used to find rules in a Dataset to satisfy user defined
// goals
package rhkit

import (
	"errors"
	"github.com/lawrencewoodman/ddataset"
	"github.com/vlifesystems/rhkit/aggregator"
	"github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/rule"
)

// ErrNoRulesGenerated indicates that no rules were generated
var ErrNoRulesGenerated = errors.New("no rules generated")

// DescribeError indicates an error describing a Dataset
type DescribeError struct {
	Err error
}

func (e DescribeError) Error() string {
	return "problem describing dataset: " + e.Err.Error()
}

// GenerateRulesError indicates an error generating rules
type GenerateRulesError struct {
	Err error
}

func (e GenerateRulesError) Error() string {
	return "problem generating rules: " + e.Err.Error()
}

// AssessError indicates an error assessing rules
type AssessError struct {
	Err error
}

func (e AssessError) Error() string {
	return "problem assessing rules: " + e.Err.Error()
}

type Options struct {
	MaxNumRules             int
	RuleFields              []string
	GenerateArithmeticRules bool
}

func (o Options) Fields() []string {
	return o.RuleFields
}

func (o Options) Arithmetic() bool {
	return o.GenerateArithmeticRules
}

// Process processes a Dataset to find Rules to meet the supplied requirements
func Process(
	dataset ddataset.Dataset,
	aggregators []aggregator.Spec,
	goals []*goal.Goal,
	sortOrder []assessment.SortOrder,
	rules []rule.Rule,
	opts Options,
) (*assessment.Assessment, error) {
	fieldDescriptions, err := description.DescribeDataset(dataset)
	if err != nil {
		return nil, DescribeError{Err: err}
	}
	if len(opts.RuleFields) == 0 {
		rules = append(rules, rule.NewTrue())
	}
	ass := assessment.New(aggregators, goals)
	if err := ass.AssessRules(dataset, rules); err != nil {
		return nil, AssessError{Err: err}
	}

	if len(opts.RuleFields) > 0 {
		err := processGenerate(
			ass,
			dataset,
			fieldDescriptions,
			sortOrder,
			len(rules),
			opts,
		)
		if err != nil {
			return nil, err
		}
	}
	ass.Sort(sortOrder)
	ass.Refine()

	if opts.MaxNumRules-len(rules) < 1 {
		return ass.TruncateRuleAssessments(1), nil
	}
	return ass.TruncateRuleAssessments(opts.MaxNumRules - len(rules)), nil
}

func processGenerate(
	ass *assessment.Assessment,
	dataset ddataset.Dataset,
	fieldDescriptions *description.Description,
	sortOrder []assessment.SortOrder,
	numUserRules int,
	opts Options,
) error {
	generatedRules, err := rule.Generate(fieldDescriptions, opts)
	if err != nil {
		return GenerateRulesError{Err: err}
	}
	if len(generatedRules) < 2 {
		return ErrNoRulesGenerated
	}

	if err := ass.AssessRules(dataset, generatedRules); err != nil {
		return AssessError{Err: err}
	}

	ass.Sort(sortOrder)
	ass.Refine()

	if len(opts.Fields()) == 2 {
		cRules := rule.Combine(ass.Rules(), 5000)
		if err := ass.AssessRules(dataset, cRules); err != nil {
			return AssessError{Err: err}
		}
		ass.Sort(sortOrder)
		ass.Refine()
	}

	bestRules := ass.Rules()

	tweakableRules := rule.Tweak(1, bestRules, fieldDescriptions)
	if err := ass.AssessRules(dataset, tweakableRules); err != nil {
		return AssessError{Err: err}
	}
	ass.Sort(sortOrder)
	ass.Refine()

	bestRules = ass.Rules()
	reducedDPRules := rule.ReduceDP(bestRules)

	if err := ass.AssessRules(dataset, reducedDPRules); err != nil {
		return AssessError{Err: err}
	}
	ass.Sort(sortOrder)
	ass.Refine()

	combinedRules := rule.Combine(ass.Rules(), 2000)

	if err := ass.AssessRules(dataset, combinedRules); err != nil {
		return AssessError{Err: err}
	}
	return nil
}
