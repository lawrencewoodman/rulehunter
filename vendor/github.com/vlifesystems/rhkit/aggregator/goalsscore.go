// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package aggregator

import (
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/goal"
)

type goalsScoreAggregator struct{}

type goalsScoreSpec struct {
	name string
}

type goalsScoreInstance struct {
	spec *goalsScoreSpec
}

func init() {
	Register("goalsscore", &goalsScoreAggregator{})
}

func (a *goalsScoreAggregator) MakeSpec(
	name string,
	expr string,
) (Spec, error) {
	d := &goalsScoreSpec{name: name}
	return d, nil
}

func (ad *goalsScoreSpec) New() Instance {
	return &goalsScoreInstance{spec: ad}
}

func (ad *goalsScoreSpec) Name() string {
	return ad.name
}

func (ad *goalsScoreSpec) Kind() string {
	return "goalsscore"
}

func (ad *goalsScoreSpec) Arg() string {
	return ""
}

func (ai *goalsScoreInstance) Name() string {
	return ai.spec.name
}

func (ai *goalsScoreInstance) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	return nil
}

func (ai *goalsScoreInstance) Result(
	aggregatorInstances []Instance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	instancesMap, err :=
		InstancesToMap(aggregatorInstances, goals, numRecords, ai.Name())
	if err != nil {
		return dlit.MustNew(err)
	}
	goalsScore := 0.0
	increment := 1.0
	for _, goal := range goals {
		hasPassed, err := goal.Assess(instancesMap)
		if err != nil {
			return dlit.MustNew(err)
		}

		if hasPassed {
			goalsScore += increment
		} else {
			increment = 0.001
		}
	}
	return dlit.MustNew(goalsScore)
}
