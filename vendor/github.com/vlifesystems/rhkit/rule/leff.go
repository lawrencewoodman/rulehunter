// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/vlifesystems/rhkit/description"
)

// LEFF represents a rule determining if fieldA <= fieldB
type LEFF struct {
	fieldA string
	fieldB string
}

func init() {
	registerGenerator("LEFF", generateLEFF)
}

func NewLEFF(fieldA, fieldB string) Rule {
	return &LEFF{fieldA: fieldA, fieldB: fieldB}
}

func (r *LEFF) String() string {
	return r.fieldA + " <= " + r.fieldB
}

func (r *LEFF) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.fieldA]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}
	rh, ok := record[r.fieldB]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return lhInt <= rhInt, nil
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsFloat {
		return lhFloat <= rhFloat, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *LEFF) Fields() []string {
	return []string{r.fieldA, r.fieldB}
}

func generateLEFF(
	inputDescription *description.Description,
	generationDesc GenerationDescriber,
) []Rule {
	rules := make([]Rule, 0)
	for _, field := range generationDesc.Fields() {
		fd := inputDescription.Fields[field]
		if fd.Kind != description.Number {
			continue
		}
		fieldNum := description.CalcFieldNum(inputDescription.Fields, field)
		for _, oField := range generationDesc.Fields() {
			oFd := inputDescription.Fields[oField]
			oFieldNum := description.CalcFieldNum(inputDescription.Fields, oField)
			isComparable := hasComparableNumberRange(fd, oFd)
			if fieldNum < oFieldNum && isComparable {
				r := NewLEFF(field, oField)
				rules = append(rules, r)
			}
		}
	}
	return rules
}
