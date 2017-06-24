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

func LoadJSON(filename string) (*Description, error) {
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
	panic("can't find field in Field descriptions: " + fieldN)
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

func (d *Description) CheckEqual(o *Description) error {
	if len(d.Fields) != len(o.Fields) {
		return fmt.Errorf(
			"number of Fields doesn't match: %d != %d",
			len(d.Fields), len(o.Fields),
		)
	}
	for field, fd := range d.Fields {
		oFd, ok := o.Fields[field]
		if !ok {
			return fmt.Errorf("missing field: %s", field)
		}
		if err := fd.checkEqual(oFd); err != nil {
			return fmt.Errorf("description for field: %s, %s", field, err)
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

	if f.Kind == fieldtype.Number {
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

func fieldValuesEqual(
	valuesA map[string]Value,
	valuesB map[string]Value,
) error {
	if len(valuesA) != len(valuesB) {
		return fmt.Errorf("number of Values not equal: %d != %d",
			len(valuesA), len(valuesB))
	}
	for k, vA := range valuesA {
		vB, ok := valuesB[k]
		if !ok {
			return fmt.Errorf("Value missing: %s", k)
		}
		if vA.Num != vB.Num || vA.Value.String() != vB.Value.String() {
			return fmt.Errorf("Value not equal for: %s, %v != %v", k, vA, vB)
		}
	}
	return nil
}
