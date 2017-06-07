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

package description

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
	"github.com/vlifesystems/rhkit/internal/fieldtype"
	"io/ioutil"
	"os"
	"sort"
)

type Description struct {
	Fields map[string]*Field
}

type Field struct {
	Kind      fieldtype.FieldType
	Min       *dlit.Literal
	Max       *dlit.Literal
	MaxDP     int
	Values    map[string]Value
	NumValues int
}

type Value struct {
	Value *dlit.Literal
	Num   int
}

func LoadDescriptionJSON(filename string) (*Description, error) {
	var dj descriptionJ

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&dj); err != nil {
		return nil, err
	}

	fields := make(map[string]*Field, len(dj.Fields))
	for field, fd := range dj.Fields {
		values := make(map[string]Value, len(fd.Values))
		for v, vd := range fd.Values {
			values[v] = Value{
				Value: dlit.NewString(vd.Value),
				Num:   vd.Num,
			}
		}
		fields[field] = &Field{
			Kind:      fieldtype.New(fd.Kind),
			Min:       dlit.NewString(fd.Min),
			Max:       dlit.NewString(fd.Max),
			MaxDP:     fd.MaxDP,
			Values:    values,
			NumValues: fd.NumValues,
		}
	}
	d := &Description{Fields: fields}
	return d, nil
}

// Calculates the field number based on the string sorted order of
// the field names
func CalcFieldNum(fieldDescriptions map[string]*Field, fieldN string) int {
	fields := make([]string, len(fieldDescriptions))
	i := 0
	for field := range fieldDescriptions {
		fields[i] = field
		i++
	}
	sort.Strings(fields)
	j := 0
	for _, field := range fields {
		if field == fieldN {
			return j
		}
		j++
	}
	panic("can't find field in fieldDescriptions")
}

func (d *Description) WriteJSON(filename string) error {
	fields := make(map[string]*fieldJ, len(d.Fields))
	for field, fd := range d.Fields {
		fields[field] = newFieldDescriptionJ(fd)
	}
	dj := descriptionJ{Fields: fields}
	json, err := json.Marshal(dj)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, json, 0640)
}

func (d *Description) CheckEqual(dWant *Description) error {
	if len(d.Fields) != len(dWant.Fields) {
		return fmt.Errorf(
			"number of FieldDescriptions doesn't match. got: %d, want: %d\n",
			len(d.Fields), len(dWant.Fields),
		)
	}
	for field, fdG := range d.Fields {
		fdW, ok := dWant.Fields[field]
		if !ok {
			return fmt.Errorf("field Description missing for field: %s", field)
		}
		if err := fdG.checkEqual(fdW); err != nil {
			return fmt.Errorf("field Description for field: %s, %s", field, err)
		}
	}
	return nil
}

// Create a New Description.
func New() *Description {
	fd := map[string]*Field{}
	return &Description{fd}
}

type descriptionJ struct {
	Fields map[string]*fieldJ
}

type fieldJ struct {
	Kind      string
	Min       string
	Max       string
	MaxDP     int
	Values    map[string]valueDescriptionJ
	NumValues int
}

type valueDescriptionJ struct {
	Value string
	Num   int
}

func newFieldDescriptionJ(fd *Field) *fieldJ {
	values := make(map[string]valueDescriptionJ, len(fd.Values))
	for v, vd := range fd.Values {
		values[v] = valueDescriptionJ{
			Value: vd.Value.String(),
			Num:   vd.Num,
		}
	}
	min := ""
	max := ""
	if fd.Min != nil {
		min = fd.Min.String()
	}
	if fd.Max != nil {
		max = fd.Max.String()
	}
	return &fieldJ{
		Kind:      fd.Kind.String(),
		Min:       min,
		Max:       max,
		MaxDP:     fd.MaxDP,
		Values:    values,
		NumValues: fd.NumValues,
	}
}

func (fd *Field) String() string {
	return fmt.Sprintf("Kind: %s, Min: %s, Max: %s, MaxDP: %d, Values: %v",
		fd.Kind, fd.Min, fd.Max, fd.MaxDP, fd.Values)
}

// Analyse this record
func (d *Description) NextRecord(record ddataset.Record) {
	if len(d.Fields) == 0 {
		for field, value := range record {
			d.Fields[field] = &Field{
				Kind:   fieldtype.Unknown,
				Min:    value,
				Max:    value,
				Values: map[string]Value{},
			}
		}
	}

	for field, value := range record {
		d.Fields[field].processValue(value)
	}
}

func (f *Field) processValue(value *dlit.Literal) {
	f.updateKind(value)
	f.updateValues(value)
	f.updateNumBoundaries(value)
}

func (f *Field) updateKind(value *dlit.Literal) {
	switch f.Kind {
	case fieldtype.Unknown:
		fallthrough
	case fieldtype.Number:
		if _, isInt := value.Int(); isInt {
			f.Kind = fieldtype.Number
			break
		}
		if _, isFloat := value.Float(); isFloat {
			f.Kind = fieldtype.Number
			break
		}
		f.Kind = fieldtype.String
	}
}

func (f *Field) updateValues(value *dlit.Literal) {
	// Chose 31 so could hold each day in month
	const maxNumValues = 31
	if f.Kind == fieldtype.Ignore ||
		f.Kind == fieldtype.Unknown ||
		f.NumValues == -1 {
		return
	}
	if vd, ok := f.Values[value.String()]; ok {
		f.Values[value.String()] = Value{vd.Value, vd.Num + 1}
		return
	}
	if f.NumValues >= maxNumValues {
		if f.Kind == fieldtype.String {
			f.Kind = fieldtype.Ignore
		}
		f.Values = map[string]Value{}
		f.NumValues = -1
		return
	}
	f.NumValues++
	f.Values[value.String()] = Value{value, 1}
}

func (f *Field) updateNumBoundaries(value *dlit.Literal) {
	if f.Kind == fieldtype.Number {
		vars := map[string]*dlit.Literal{"min": f.Min, "max": f.Max, "v": value}
		f.Min = dexpr.Eval("min(min, v)", dexprfuncs.CallFuncs, vars)
		f.Max = dexpr.Eval("max(max, v)", dexprfuncs.CallFuncs, vars)
		f.MaxDP =
			int(maxI(int64(f.MaxDP), int64(internal.NumDecPlaces(value.String()))))
	}
}

func (f *Field) checkEqual(fdWant *Field) error {
	if f.Kind != fdWant.Kind {
		return fmt.Errorf("got field kind: %s, want: %s", f.Kind, fdWant.Kind)
	}
	if len(f.Values) != len(fdWant.Values) {
		return fmt.Errorf("got %d values, want: %d",
			len(f.Values), len(fdWant.Values))
	}
	if f.Kind == fieldtype.Number {
		if f.Min.String() != fdWant.Min.String() ||
			f.Max.String() != fdWant.Max.String() {
			return fmt.Errorf("got min: %s and max: %s, want min: %s and max: %s",
				f.Min, f.Max, fdWant.Min, fdWant.Max)
		}
		if f.MaxDP != fdWant.MaxDP {
			return fmt.Errorf("got maxDP: %d, want: %d", f.MaxDP, fdWant.MaxDP)
		}
	}

	if f.NumValues != fdWant.NumValues {
		return fmt.Errorf("got numValues: %d, numValues: %d",
			f.NumValues, fdWant.NumValues)
	}

	return fieldValuesEqual(f.Values, fdWant.Values)
}

func fieldValuesEqual(
	vdsGot map[string]Value,
	vdsWant map[string]Value,
) error {
	if len(vdsGot) != len(vdsWant) {
		return fmt.Errorf("got %d valueDescriptions, want: %d",
			len(vdsGot), len(vdsWant))
	}
	for k, vdW := range vdsWant {
		vdG, ok := vdsGot[k]
		if !ok {
			return fmt.Errorf("valueDescription missing value: %s", k)
		}
		if vdG.Num != vdW.Num || vdG.Value.String() != vdW.Value.String() {
			return fmt.Errorf("got valueDescription: %v, want: %v", vdG, vdW)
		}
	}
	return nil
}

func maxI(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
