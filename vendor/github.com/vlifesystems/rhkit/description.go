/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
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

package rhkit

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
	"io/ioutil"
	"math"
	"os"
)

type Description struct {
	Fields map[string]*fieldDescription
}

type fieldDescription struct {
	Kind      fieldType
	Min       *dlit.Literal
	Max       *dlit.Literal
	MaxDP     int
	Values    map[string]valueDescription
	NumValues int
}

type valueDescription struct {
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

	fields := make(map[string]*fieldDescription, len(dj.Fields))
	for field, fd := range dj.Fields {
		values := make(map[string]valueDescription, len(fd.Values))
		for v, vd := range fd.Values {
			values[v] = valueDescription{
				Value: dlit.NewString(vd.Value),
				Num:   vd.Num,
			}
		}
		fields[field] = &fieldDescription{
			Kind:      newFieldType(fd.Kind),
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

func (d *Description) WriteJSON(filename string) error {
	fields := make(map[string]*fieldDescriptionJ, len(d.Fields))
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

// Create a New Description.
func newDescription() *Description {
	fd := map[string]*fieldDescription{}
	return &Description{fd}
}

type descriptionJ struct {
	Fields map[string]*fieldDescriptionJ
}

type fieldDescriptionJ struct {
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

func newFieldDescriptionJ(fd *fieldDescription) *fieldDescriptionJ {
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
	return &fieldDescriptionJ{
		Kind:      fd.Kind.String(),
		Min:       min,
		Max:       max,
		MaxDP:     fd.MaxDP,
		Values:    values,
		NumValues: fd.NumValues,
	}
}

func (fd *fieldDescription) String() string {
	return fmt.Sprintf("Kind: %s, Min: %s, Max: %s, MaxDP: %d, Values: %s",
		fd.Kind, fd.Min, fd.Max, fd.MaxDP, fd.Values)
}

// Analyse this record
func (d *Description) NextRecord(record ddataset.Record) {
	if len(d.Fields) == 0 {
		for field, value := range record {
			d.Fields[field] = &fieldDescription{
				Kind:   ftUnknown,
				Min:    value,
				Max:    value,
				Values: map[string]valueDescription{},
			}
		}
	}

	for field, value := range record {
		d.Fields[field].processValue(value)
	}
}

func (f *fieldDescription) processValue(value *dlit.Literal) {
	f.updateKind(value)
	f.updateValues(value)
	f.updateNumBoundaries(value)
}

func (f *fieldDescription) updateKind(value *dlit.Literal) {
	switch f.Kind {
	case ftUnknown:
		fallthrough
	case ftInt:
		if _, isInt := value.Int(); isInt {
			f.Kind = ftInt
			break
		}
		fallthrough
	case ftFloat:
		if _, isFloat := value.Float(); isFloat {
			f.Kind = ftFloat
			break
		}
		f.Kind = ftString
	}
}

func (f *fieldDescription) updateValues(value *dlit.Literal) {
	// Chose 31 so could hold each day in month
	const maxNumValues = 31
	if f.Kind == ftIgnore ||
		f.Kind == ftUnknown ||
		f.NumValues == -1 {
		return
	}
	if vd, ok := f.Values[value.String()]; ok {
		f.Values[value.String()] = valueDescription{vd.Value, vd.Num + 1}
		return
	}
	if f.NumValues >= maxNumValues {
		if f.Kind == ftString {
			f.Kind = ftIgnore
		}
		f.Values = map[string]valueDescription{}
		f.NumValues = -1
		return
	}
	f.NumValues++
	f.Values[value.String()] = valueDescription{value, 1}
}

func (f *fieldDescription) updateNumBoundaries(value *dlit.Literal) {
	if f.Kind == ftInt {
		valueInt, valueIsInt := value.Int()
		minInt, minIsInt := f.Min.Int()
		maxInt, maxIsInt := f.Max.Int()
		if !valueIsInt || !minIsInt || !maxIsInt {
			panic("Type mismatch")
		}
		f.Min = dlit.MustNew(minI(minInt, valueInt))
		f.Max = dlit.MustNew(maxI(maxInt, valueInt))
	} else if f.Kind == ftFloat {
		valueFloat, valueIsFloat := value.Float()
		minFloat, minIsFloat := f.Min.Float()
		maxFloat, maxIsFloat := f.Max.Float()
		if !valueIsFloat || !minIsFloat || !maxIsFloat {
			panic("Type mismatch")
		}
		f.Min = dlit.MustNew(math.Min(minFloat, valueFloat))
		f.Max = dlit.MustNew(math.Max(maxFloat, valueFloat))
		f.MaxDP =
			int(maxI(int64(f.MaxDP), int64(numDecPlaces(value.String()))))
	}
}

func minI(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func maxI(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
