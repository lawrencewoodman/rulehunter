// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/vlifesystems/rhkit/description"
)

// EQFF represents a rule determining if fieldA == fieldB
type EQFF struct {
	fieldA string
	fieldB string
}

func init() {
	registerGenerator("EQFF", generateEQFF)
}

func NewEQFF(fieldA, fieldB string) Rule {
	return &EQFF{fieldA: fieldA, fieldB: fieldB}
}

func (r *EQFF) String() string {
	return r.fieldA + " == " + r.fieldB
}

func (r *EQFF) IsTrue(record ddataset.Record) (bool, error) {
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
		return lhInt == rhInt, nil
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsFloat {
		return lhFloat == rhFloat, nil
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools
	lhErr := lh.Err()
	rhErr := rh.Err()
	if lhErr != nil || rhErr != nil {
		return false, IncompatibleTypesRuleError{Rule: r}
	}

	return lh.String() == rh.String(), nil
}

func (r *EQFF) Fields() []string {
	return []string{r.fieldA, r.fieldB}
}

func generateEQFF(
	inputDescription *description.Description,
	generationDesc GenerationDescriber,
) []Rule {
	rules := make([]Rule, 0)
	for _, field := range generationDesc.Fields() {
		fd := inputDescription.Fields[field]
		if fd.Kind != description.String && fd.Kind != description.Number {
			continue
		}
		fieldNum := description.CalcFieldNum(inputDescription.Fields, field)
		for _, oField := range generationDesc.Fields() {
			oFd := inputDescription.Fields[oField]
			if oFd.Kind == fd.Kind {
				oFieldNum := description.CalcFieldNum(inputDescription.Fields, oField)
				numSharedValues := calcNumSharedValues(fd, oFd)
				if fieldNum < oFieldNum && numSharedValues >= 2 {
					r := NewEQFF(field, oField)
					rules = append(rules, r)
				}
			}
		}
	}
	return rules
}
