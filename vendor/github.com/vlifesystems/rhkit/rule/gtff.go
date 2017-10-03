// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/internal/fieldtype"
)

// GTFF represents a rule determining if fieldA > fieldB
type GTFF struct {
	fieldA string
	fieldB string
}

func init() {
	registerGenerator("GTFF", generateGTFF)
}

func NewGTFF(fieldA, fieldB string) Rule {
	return &GTFF{fieldA: fieldA, fieldB: fieldB}
}

func (r *GTFF) String() string {
	return r.fieldA + " > " + r.fieldB
}

func (r *GTFF) IsTrue(record ddataset.Record) (bool, error) {
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
		return lhInt > rhInt, nil
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsFloat {
		return lhFloat > rhFloat, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *GTFF) Fields() []string {
	return []string{r.fieldA, r.fieldB}
}

func generateGTFF(
	inputDescription *description.Description,
	generationDesc GenerationDescriber,
	field string,
) []Rule {
	fd := inputDescription.Fields[field]
	if fd.Kind != fieldtype.Number {
		return []Rule{}
	}
	fieldNum := description.CalcFieldNum(inputDescription.Fields, field)
	rules := make([]Rule, 0)
	for oField, oFd := range inputDescription.Fields {
		oFieldNum := description.CalcFieldNum(inputDescription.Fields, oField)
		isComparable := hasComparableNumberRange(fd, oFd)
		if fieldNum < oFieldNum && isComparable &&
			internal.IsStringInSlice(oField, generationDesc.Fields()) {
			r := NewGTFF(field, oField)
			rules = append(rules, r)
		}
	}
	return rules
}
