/*
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of rhkit.

	rhkit is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	rhkit is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with rhkit; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

package rhkit

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
	assessment, err := newAssessment(numRecords, goodRuleAssessors, e.Goals)
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
