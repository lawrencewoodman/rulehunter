// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/description"
)

// NEFV represents a rule determining if fieldA != value
type NEFV struct {
	field string
	value *dlit.Literal
}

func init() {
	registerGenerator("NEFV", generateNEFV)
}

func NewNEFV(field string, value *dlit.Literal) Rule {
	return &NEFV{field: field, value: value}
}

func (r *NEFV) String() string {
	_, vIsFloat := r.value.Float()
	_, vIsInt := r.value.Int()
	if vIsInt || vIsFloat {
		return fmt.Sprintf("%s != %s", r.field, r.value)
	}
	return fmt.Sprintf("%s != \"%s\"", r.field, r.value)
}

func (r *NEFV) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}
	if lhInt, lhIsInt := lh.Int(); lhIsInt {
		if vInt, vIsInt := r.value.Int(); vIsInt {
			return lhInt != vInt, nil
		}
	}
	if lhFloat, lhIsFloat := lh.Float(); lhIsFloat {
		if vFloat, vIsFloat := r.value.Float(); vIsFloat {
			return lhFloat != vFloat, nil
		}
	}
	if lh.Err() == nil && r.value.Err() == nil {
		return lh.String() != r.value.String(), nil
	}
	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *NEFV) Fields() []string {
	return []string{r.field}
}

func generateNEFV(
	inputDescription *description.Description,
	generationDesc GenerationDescriber,
) []Rule {
	rules := make([]Rule, 0)
	for _, field := range generationDesc.Fields() {
		fd := inputDescription.Fields[field]
		values := fd.Values
		if len(values) > 2 && fd.Kind != description.Ignore {
			for _, vd := range values {
				if vd.Num >= 2 {
					r := NewNEFV(field, vd.Value)
					rules = append(rules, r)
				}
			}
		}
	}
	return rules
}
