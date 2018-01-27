// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// Package aggregator describes and handles Aggregators
package aggregator

import (
	"sync"

	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

var (
	aggregatorsMu sync.RWMutex
	aggregators   = make(map[string]Aggregator)
)

type Aggregator interface {
	MakeSpec(string, string) (Spec, error)
}

type Desc struct {
	Name string
	Kind string
	Arg  string
}

type Spec interface {
	New() Instance
	Name() string
	Kind() string
	Arg() string
}

type Instance interface {
	Name() string
	Result([]Instance, []*goal.Goal, int64) *dlit.Literal
	NextRecord(map[string]*dlit.Literal, bool) error
}

// Register makes an Aggregator available by the provided kind.
// If Register is called twice with the same kind or if
// aggregator is nil, it panics.
func Register(kind string, aggregator Aggregator) {
	aggregatorsMu.Lock()
	defer aggregatorsMu.Unlock()
	if aggregator == nil {
		panic("aggregator.Register aggregator is nil")
	}
	if _, dup := aggregators[kind]; dup {
		panic("aggregator.Register called twice for aggregator: " + kind)
	}
	aggregators[kind] = aggregator
}

// Create a new Aggregator where 'name' is what the aggregator will be
// known as, 'kind' is the name of the Aggregator as Registered,
// 'args' are any arguments to pass to the Aggregator.
func New(name string, kind string, args ...string) (Spec, error) {
	var spec Spec
	var err error
	aggregatorsMu.RLock()
	aggregator, ok := aggregators[kind]
	aggregatorsMu.RUnlock()
	if !ok {
		return nil, DescError{Name: name, Kind: kind, Err: ErrUnregisteredKind}
	}

	if !internal.IsIdentifierValid(name) {
		return nil, DescError{Name: name, Kind: kind, Err: ErrInvalidName}
	}

	if kind == "goalsscore" {
		if len(args) != 0 {
			return nil, DescError{Name: name, Kind: kind, Err: ErrInvalidNumArgs}
		}
		spec, err = aggregator.MakeSpec(name, "")
	} else {
		if len(args) != 1 {
			return nil, DescError{Name: name, Kind: kind, Err: ErrInvalidNumArgs}
		}
		spec, err = aggregator.MakeSpec(name, args[0])
	}

	if err != nil {
		return nil, DescError{Name: name, Kind: kind, Err: err}
	}
	return spec, nil
}

func MustNew(name string, kind string, args ...string) Spec {
	a, err := New(name, kind, args...)
	if err != nil {
		panic(err)
	}
	return a
}

// InstancesToMap gets the results of each Instance and
// returns the results as a map with the aggregatorSpec name as the key
func InstancesToMap(
	Instances []Instance,
	goals []*goal.Goal,
	numRecords int64,
	stopNames ...string,
) (map[string]*dlit.Literal, error) {
	r := make(map[string]*dlit.Literal, len(Instances))
	numRecordsL := dlit.MustNew(numRecords)
	r["numRecords"] = numRecordsL
	for _, ai := range Instances {
		for _, stopName := range stopNames {
			if stopName == ai.Name() {
				return r, nil
			}
		}
		l := ai.Result(Instances, goals, numRecords)
		if err := l.Err(); err != nil {
			return r, err
		}
		r[ai.Name()] = l
	}
	return r, nil
}

func MakeSpecs(
	fields []string,
	descs []*Desc,
) ([]Spec, error) {
	var err error
	r := make([]Spec, len(descs))
	for i, desc := range descs {
		if err = checkDescValid(fields, desc); err != nil {
			return []Spec{}, err
		}
		r[i], err = New(desc.Name, desc.Kind, desc.Arg)
		if err != nil {
			return []Spec{}, err
		}
	}
	return addDefaultAggregators(r), nil
}

func addDefaultAggregators(specs []Spec) []Spec {
	newSpecs := make([]Spec, 2)
	newSpecs[0] = MustNew("numMatches", "count", "true()")
	newSpecs[1] = MustNew(
		"percentMatches",
		"calc",
		"roundto(100.0 * numMatches / numRecords, 2)",
	)
	goalsScoreSpec := MustNew("goalsScore", "goalsscore")
	newSpecs = append(newSpecs, specs...)
	newSpecs = append(newSpecs, goalsScoreSpec)
	return newSpecs
}

func checkDescValid(fields []string, desc *Desc) error {
	if internal.IsStringInSlice(desc.Name, fields) {
		return DescError{Name: desc.Name, Kind: desc.Kind, Err: ErrNameClash}
	}
	if desc.Name == "percentMatches" ||
		desc.Name == "numMatches" ||
		desc.Name == "goalsScore" {
		return DescError{Name: desc.Name, Kind: desc.Kind, Err: ErrNameReserved}
	}
	return nil
}

func roundTo(l *dlit.Literal, dp int) *dlit.Literal {
	var roundExpr = dexpr.MustNew("roundto(n, dp)", dexprfuncs.CallFuncs)
	vars := map[string]*dlit.Literal{
		"n":  l,
		"dp": dlit.MustNew(dp),
	}
	return roundExpr.Eval(vars)
}
