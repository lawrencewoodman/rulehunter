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

// Package aggregators describes and handles Aggregators
package aggregators

import (
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/goal"
	"sync"
)

var (
	aggregatorsMu sync.RWMutex
	aggregators   = make(map[string]Aggregator)
)

type Aggregator interface {
	MakeSpec(string, string) (AggregatorSpec, error)
}

type AggregatorSpec interface {
	New() AggregatorInstance
	Name() string
	Kind() string
	Arg() string
}

type AggregatorInstance interface {
	Name() string
	Result([]AggregatorInstance, []*goal.Goal, int64) *dlit.Literal
	NextRecord(map[string]*dlit.Literal, bool) error
}

// Register makes an Aggregator available by the provided aggType.
// If Register is called twice with the same aggType or if
// aggregator is nil, it panics.
func Register(aggType string, aggregator Aggregator) {
	aggregatorsMu.Lock()
	defer aggregatorsMu.Unlock()
	if aggregator == nil {
		panic("aggregator.Register aggregator is nil")
	}
	if _, dup := aggregators[aggType]; dup {
		panic("aggregator.Register called twice for aggregator: " + aggType)
	}
	aggregators[aggType] = aggregator
}

// Create a new Aggregator where 'name' is what the aggregator will be
// known as, 'aggType' is the name of the Aggregator as Registered,
// 'args' are any arguments to pass to the Aggregator.
func New(name string, aggType string, args ...string) (AggregatorSpec, error) {
	var ad AggregatorSpec
	var err error
	aggregatorsMu.RLock()
	aggregator, ok := aggregators[aggType]
	aggregatorsMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unrecognized aggregator: %s", aggType)
	}

	if aggType == "goalsscore" {
		if len(args) != 0 {
			return nil,
				fmt.Errorf("invalid number of arguments for aggregator: goalsscore")
		}
		ad, err = aggregator.MakeSpec(name, "")
	} else {
		if len(args) != 1 {
			return nil,
				fmt.Errorf("invalid number of arguments for aggregator: %s", aggType)
		}
		ad, err = aggregator.MakeSpec(name, args[0])
	}

	if err != nil {
		return nil,
			fmt.Errorf("can't make aggregator: %s, error: %s", name, err)
	}
	return ad, nil
}

func MustNew(name string, aggType string, args ...string) AggregatorSpec {
	a, err := New(name, aggType, args...)
	if err != nil {
		panic(err)
	}
	return a
}

// InstancesToMap gets the results of each AggregatorInstance and
// returns the results as a map with the aggregatorSpec name as the key
func InstancesToMap(
	aggregatorInstances []AggregatorInstance,
	goals []*goal.Goal,
	numRecords int64,
	stopNames ...string,
) (map[string]*dlit.Literal, error) {
	r := make(map[string]*dlit.Literal, len(aggregatorInstances))
	numRecordsL := dlit.MustNew(numRecords)
	r["numRecords"] = numRecordsL
	for _, ai := range aggregatorInstances {
		for _, stopName := range stopNames {
			if stopName == ai.Name() {
				return r, nil
			}
		}
		l := ai.Result(aggregatorInstances, goals, numRecords)
		if err := l.Err(); err != nil {
			return r, err
		}
		r[ai.Name()] = l
	}
	return r, nil
}
