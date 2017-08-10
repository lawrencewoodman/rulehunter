// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package assessment

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/aggregator"
	"strings"
)

type SortDesc struct {
	Aggregator string
	Direction  string
}

type SortOrder struct {
	Aggregator string
	Direction  direction
}

type direction int

const (
	ASCENDING direction = iota
	DESCENDING
)

type SortOrderError struct {
	Aggregator string
	Direction  string
	Err        error
}

var (
	ErrInvalidDirection       = errors.New("invalid direction")
	ErrUnrecognisedAggregator = errors.New("unrecognised aggregator")
)

func (e SortOrderError) Error() string {
	return fmt.Sprintf(
		"problem with sort order - aggregator: %s, direction: %s (%s)",
		e.Aggregator, e.Direction, e.Err,
	)
}

type Namer interface {
	Name() string
}

func NewSortOrder(
	aggregatorSpecs []aggregator.Spec,
	aggregator string,
	direction string,
) (SortOrder, error) {
	for _, s := range aggregatorSpecs {
		if s.Name() == aggregator {
			// TODO: Make case insensitive?
			switch direction {
			case "ascending":
				return SortOrder{aggregator, ASCENDING}, nil
			case "descending":
				return SortOrder{aggregator, DESCENDING}, nil
			}
			return SortOrder{}, SortOrderError{
				Aggregator: aggregator,
				Direction:  direction,
				Err:        ErrInvalidDirection,
			}
		}
	}
	return SortOrder{}, SortOrderError{
		Aggregator: aggregator,
		Direction:  direction,
		Err:        ErrUnrecognisedAggregator,
	}
}

func MakeSortOrders(
	aggregatorSpecs []aggregator.Spec,
	descs []SortDesc,
) ([]SortOrder, error) {
	var err error
	r := make([]SortOrder, len(descs))
	for i, desc := range descs {
		r[i], err = NewSortOrder(
			aggregatorSpecs,
			desc.Aggregator,
			desc.Direction,
		)
		if err != nil {
			return []SortOrder{}, err
		}
	}
	return r, nil
}

func (d direction) String() string {
	if d == ASCENDING {
		return "ascending"
	}
	return "descending"
}

// by implements sort.Interface for []*RuleAssessments based
// on the sortFields
type by struct {
	ruleAssessments []*RuleAssessment
	sortOrders      []SortOrder
}

func (b by) Len() int { return len(b.ruleAssessments) }
func (b by) Swap(i, j int) {
	b.ruleAssessments[i], b.ruleAssessments[j] =
		b.ruleAssessments[j], b.ruleAssessments[i]
}

func (b by) Less(i, j int) bool {
	var vI *dlit.Literal
	var vJ *dlit.Literal
	for _, sortOrder := range b.sortOrders {
		aggregator := sortOrder.Aggregator
		direction := sortOrder.Direction
		vI = b.ruleAssessments[i].Aggregators[aggregator]
		vJ = b.ruleAssessments[j].Aggregators[aggregator]
		c := compareDlitNums(vI, vJ)

		if direction == DESCENDING {
			c *= -1
		}
		if c < 0 {
			return true
		} else if c > 0 {
			return false
		}
	}

	ruleStrI := b.ruleAssessments[i].Rule.String()
	ruleStrJ := b.ruleAssessments[j].Rule.String()
	ruleLenI := len(ruleStrI)
	ruleLenJ := len(ruleStrJ)
	if ruleLenI != ruleLenJ {
		return ruleLenI < ruleLenJ
	}

	return strings.Compare(ruleStrI, ruleStrJ) == -1
}

func compareDlitNums(l1 *dlit.Literal, l2 *dlit.Literal) int {
	// TODO: Use a dexpr to do this
	i1, l1IsInt := l1.Int()
	i2, l2IsInt := l2.Int()
	if l1IsInt && l2IsInt {
		if i1 < i2 {
			return -1
		}
		if i1 > i2 {
			return 1
		}
		return 0
	}

	f1, l1IsFloat := l1.Float()
	f2, l2IsFloat := l2.Float()

	if l1IsFloat && l2IsFloat {
		if f1 < f2 {
			return -1
		}
		if f1 > f2 {
			return 1
		}
		return 0
	}
	panic(fmt.Sprintf("can't compare numbers: %s, %s", l1, l2))
}
