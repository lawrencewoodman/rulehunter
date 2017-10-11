// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package description

import (
	"encoding/json"
	"fmt"

	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

// Field describes a field
type Field struct {
	Kind      FieldType
	Min       *dlit.Literal
	Max       *dlit.Literal
	MaxDP     int
	Values    map[string]Value
	NumValues int
}

// fieldJ is used for JSON Marshal/Unmarshal
type fieldJ struct {
	Kind      string         `json:"kind"`
	Min       string         `json:"min"`
	Max       string         `json:"max"`
	MaxDP     int            `json:"maxDP"`
	Values    map[string]int `json:"values"`
	NumValues int            `json:"numvalues"`
}

func (f *Field) UnmarshalJSON(b []byte) error {
	var fj fieldJ
	if err := json.Unmarshal(b, &fj); err != nil {
		return err
	}
	values := make(map[string]Value, len(fj.Values))
	for v, n := range fj.Values {
		values[v] = Value{Value: dlit.NewString(v), Num: n}
	}
	f.Kind = NewFieldType(fj.Kind)
	f.Min = dlit.NewString(fj.Min)
	f.Max = dlit.NewString(fj.Max)
	f.MaxDP = fj.MaxDP
	f.Values = values
	f.NumValues = fj.NumValues
	return nil
}

func (f *Field) MarshalJSON() ([]byte, error) {
	values := make(map[string]int, len(f.Values))
	for k, v := range f.Values {
		values[k] = v.Num
	}
	fj := &fieldJ{
		Kind:      f.Kind.String(),
		Min:       "",
		Max:       "",
		MaxDP:     f.MaxDP,
		Values:    values,
		NumValues: f.NumValues,
	}
	if f.Min != nil {
		fj.Min = f.Min.String()
	}
	if f.Max != nil {
		fj.Max = f.Max.String()
	}
	return json.Marshal(fj)
}

// String outputs a string representation of the field
func (fd *Field) String() string {
	return fmt.Sprintf("Kind: %s, Min: %s, Max: %s, MaxDP: %d, Values: %v",
		fd.Kind, fd.Min, fd.Max, fd.MaxDP, fd.Values)
}

func (f *Field) processValue(value *dlit.Literal) {
	f.updateKind(value)
	f.updateValues(value)
	f.updateNumBoundaries(value)
}

func (f *Field) updateKind(value *dlit.Literal) {
	switch f.Kind {
	case Unknown:
		fallthrough
	case Number:
		if _, isInt := value.Int(); isInt {
			f.Kind = Number
			break
		}
		if _, isFloat := value.Float(); isFloat {
			f.Kind = Number
			break
		}
		f.Kind = String
	}
}

func (f *Field) updateValues(value *dlit.Literal) {
	// Chose 31 so could hold each day in month
	const maxNumValues = 31
	if f.Kind == Ignore ||
		f.Kind == Unknown ||
		f.NumValues == -1 {
		return
	}
	if vd, ok := f.Values[value.String()]; ok {
		f.Values[value.String()] = Value{vd.Value, vd.Num + 1}
		return
	}
	if f.NumValues >= maxNumValues {
		if f.Kind == String {
			f.Kind = Ignore
		}
		f.Values = map[string]Value{}
		f.NumValues = -1
		return
	}
	f.NumValues++
	f.Values[value.String()] = Value{value, 1}
}

func (f *Field) updateNumBoundaries(value *dlit.Literal) {
	if f.Kind == Number {
		vars := map[string]*dlit.Literal{"min": f.Min, "max": f.Max, "v": value}
		f.Min = dexpr.Eval("min(min, v)", dexprfuncs.CallFuncs, vars)
		f.Max = dexpr.Eval("max(max, v)", dexprfuncs.CallFuncs, vars)
		numDP := internal.NumDecPlaces(value.String())
		if numDP > f.MaxDP {
			f.MaxDP = numDP
		}
	}
}

func (f *Field) checkEqual(o *Field) error {
	if f.Kind != o.Kind {
		return fmt.Errorf("Kind not equal: %s != %s", f.Kind, o.Kind)
	}
	if f.NumValues != o.NumValues {
		return fmt.Errorf("NumValues not equal: %d != %d", f.NumValues, o.NumValues)
	}

	if f.Kind == Number {
		if f.Min.String() != o.Min.String() {
			return fmt.Errorf("Min not equal: %s != %s", f.Min, o.Min)
		}
		if f.Max.String() != o.Max.String() {
			return fmt.Errorf("Max not equal: %s != %s", f.Max, o.Max)
		}
		if f.MaxDP != o.MaxDP {
			return fmt.Errorf("MaxDP not equal: %d != %d", f.MaxDP, o.MaxDP)
		}
	}
	return fieldValuesEqual(f.Values, o.Values)
}
