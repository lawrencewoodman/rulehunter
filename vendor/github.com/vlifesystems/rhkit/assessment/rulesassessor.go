// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package assessment

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/vlifesystems/rhkit/experiment"
	"github.com/vlifesystems/rhkit/rule"
)

// AssessRules runs the rules against the experiment and returns an
// Assessment along with any errors
func AssessRules(
	rules []rule.Rule,
	e *experiment.Experiment,
) (*Assessment, error) {
	ruleAssessors := make([]*ruleAssessor, len(rules))
	for i, rule := range rules {
		ruleAssessors[i] = newRuleAssessor(rule, e.Aggregators, e.Goals)
	}

	numRecords, err := processDataset(e.Dataset, ruleAssessors)
	if err != nil {
		return &Assessment{}, err
	}
	goodRuleAssessors := filterGoodRuleAssessors(ruleAssessors, numRecords)
	assessment := NewAssessment(numRecords)
	err = assessment.AddRuleAssessors(goodRuleAssessors)
	return assessment, err
}

func filterGoodRuleAssessors(
	ruleAssessments []*ruleAssessor,
	numRecords int64,
) []*ruleAssessor {
	goodRuleAssessors := make([]*ruleAssessor, 0)
	for _, ruleAssessment := range ruleAssessments {
		numMatches, exists :=
			ruleAssessment.AggregatorValue("numMatches", numRecords)
		if !exists {
			panic("numMatches doesn't exist in aggregators")
		}
		numMatchesInt, isInt := numMatches.Int()
		if !isInt {
			panic(fmt.Sprintf("can't cast numMatches to Int: %s", numMatches))
		}
		if numMatchesInt > 0 {
			goodRuleAssessors = append(goodRuleAssessors, ruleAssessment)
		}
	}
	return goodRuleAssessors
}

func processDataset(
	dataset ddataset.Dataset,
	ruleAssessors []*ruleAssessor,
) (int64, error) {
	numRecords := int64(0)
	conn, err := dataset.Open()
	if err != nil {
		return numRecords, err
	}
	defer conn.Close()

	for conn.Next() {
		record := conn.Read()
		numRecords++
		for _, ruleAssessor := range ruleAssessors {
			err := ruleAssessor.NextRecord(record)
			if err != nil {
				return numRecords, err
			}
		}
	}

	return numRecords, conn.Err()
}
