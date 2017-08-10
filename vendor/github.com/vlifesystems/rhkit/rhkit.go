// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
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
	MaxNumRules    int
	GenerateRules  bool
	RuleComplexity rule.Complexity
}

// Process processes a Dataset to find Rules to meet the supplied requirements
func Process(
	dataset ddataset.Dataset,
	ruleFields []string,
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
	if !opts.GenerateRules {
		rules = append(rules, rule.NewTrue())
	}
	ass := assessment.New()
	err = ass.AssessRules(dataset, rules, aggregators, goals)
	if err != nil {
		return nil, AssessError{Err: err}
	}

	if opts.GenerateRules {
		err := processGenerate(
			ass,
			dataset,
			ruleFields,
			fieldDescriptions,
			aggregators,
			goals,
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
	ruleFields []string,
	fieldDescriptions *description.Description,
	aggregators []aggregator.Spec,
	goals []*goal.Goal,
	sortOrder []assessment.SortOrder,
	numUserRules int,
	opts Options,
) error {
	generatedRules, err := rule.Generate(
		fieldDescriptions,
		ruleFields,
		opts.RuleComplexity,
	)
	if err != nil {
		return GenerateRulesError{Err: err}
	}
	if len(generatedRules) < 2 {
		return ErrNoRulesGenerated
	}

	err = ass.AssessRules(dataset, generatedRules, aggregators, goals)
	if err != nil {
		return AssessError{Err: err}
	}

	ass.Sort(sortOrder)
	ass.Refine()
	bestRules := ass.Rules()

	tweakableRules := rule.Tweak(1, bestRules, fieldDescriptions)
	err = ass.AssessRules(dataset, tweakableRules, aggregators, goals)
	if err != nil {
		return AssessError{Err: err}
	}
	ass.Sort(sortOrder)
	ass.Refine()

	bestRules = ass.Rules()
	reducedDPRules := rule.ReduceDP(bestRules)

	err = ass.AssessRules(dataset, reducedDPRules, aggregators, goals)
	if err != nil {
		return AssessError{Err: err}
	}
	ass.Sort(sortOrder)
	ass.Refine()

	numRulesToCombine := 50
	bestNonCombinedRules := ass.Rules(numRulesToCombine)
	combinedRules := rule.Combine(bestNonCombinedRules)

	err = ass.AssessRules(dataset, combinedRules, aggregators, goals)
	if err != nil {
		return AssessError{Err: err}
	}
	return nil
}
