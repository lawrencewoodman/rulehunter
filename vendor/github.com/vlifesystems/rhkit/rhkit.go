// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// package rhkit is used to find rules in a Dataset to satisfy user defined
// goals
package rhkit

import (
	"errors"
	"github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/experiment"
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

// AssessError indicates an error assessing rules
type AssessError struct {
	Err error
}

func (e AssessError) Error() string {
	return "problem assessing rules: " + e.Err.Error()
}

// MergeError indicates an error Merging assessments
type MergeError struct {
	Err error
}

func (e MergeError) Error() string {
	return "problem merging assessments: " + e.Err.Error()
}

// Process processes an Experiment and returns an assessment
func Process(
	experiment *experiment.Experiment,
	maxNumRules int,
) (*assessment.Assessment, error) {
	var ass *assessment.Assessment
	var newAss *assessment.Assessment
	var err error
	fieldDescriptions, err := description.DescribeDataset(experiment.Dataset)
	if err != nil {
		return nil, DescribeError{Err: err}
	}
	rules := rule.Generate(
		fieldDescriptions,
		experiment.RuleFields,
		experiment.RuleComplexity,
	)
	if len(rules) < 2 {
		return nil, ErrNoRulesGenerated
	}

	ass, err = assessment.AssessRules(rules, experiment)
	if err != nil {
		return nil, AssessError{Err: err}
	}

	ass.Sort(experiment.SortOrder)
	ass.Refine()
	rules = ass.Rules()

	tweakableRules := rule.Tweak(
		1,
		rules,
		fieldDescriptions,
	)

	newAss, err = assessment.AssessRules(tweakableRules, experiment)
	if err != nil {
		return nil, AssessError{Err: err}
	}

	ass, err = ass.Merge(newAss)
	if err != nil {
		return nil, MergeError{Err: err}
	}
	ass.Sort(experiment.SortOrder)
	ass.Refine()

	rules = ass.Rules()
	reducedDPRules := rule.ReduceDP(rules)

	newAss, err = assessment.AssessRules(reducedDPRules, experiment)
	if err != nil {
		return nil, AssessError{Err: err}
	}

	ass, err = ass.Merge(newAss)
	if err != nil {
		return nil, MergeError{Err: err}
	}
	ass.Sort(experiment.SortOrder)
	ass.Refine()

	numRulesToCombine := 50
	bestNonCombinedRules := ass.Rules(numRulesToCombine)
	combinedRules := rule.Combine(bestNonCombinedRules)

	newAss, err = assessment.AssessRules(combinedRules, experiment)
	if err != nil {
		return nil, AssessError{Err: err}
	}

	ass, err = ass.Merge(newAss)
	if err != nil {
		return nil, MergeError{Err: err}
	}
	ass.Sort(experiment.SortOrder)
	ass.Refine()

	return ass.TruncateRuleAssessments(maxNumRules), nil
}
