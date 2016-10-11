/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of Rulehunter.

	Rulehunter is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Rulehunter is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Rulehunter; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

package rhkit

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
	"math"
)

type Description struct {
	fields map[string]*fieldDescription
}

// Create a New Description.
func newDescription() *Description {
	fd := map[string]*fieldDescription{}
	return &Description{fd}
}

type fieldDescription struct {
	kind      fieldType
	min       *dlit.Literal
	max       *dlit.Literal
	maxDP     int
	values    []*dlit.Literal
	numValues int
}

func (fd *fieldDescription) String() string {
	return fmt.Sprintf("Kind: %s, Min: %s, Max: %s, MaxDP: %d, Values: %s",
		fd.kind, fd.min, fd.max, fd.maxDP, fd.values)
}

// Analyse this record
func (d *Description) NextRecord(record ddataset.Record) {
	if len(d.fields) == 0 {
		for field, value := range record {
			d.fields[field] = &fieldDescription{
				kind: ftUnknown,
				min:  value,
				max:  value,
			}
		}
	}

	for field, value := range record {
		d.fields[field].processValue(value)
	}
}

func (f *fieldDescription) processValue(value *dlit.Literal) {
	f.updateKind(value)
	f.updateValues(value)
	f.updateNumBoundaries(value)
}

func (f *fieldDescription) updateKind(value *dlit.Literal) {
	switch f.kind {
	case ftUnknown:
		fallthrough
	case ftInt:
		if _, isInt := value.Int(); isInt {
			f.kind = ftInt
			break
		}
		fallthrough
	case ftFloat:
		if _, isFloat := value.Float(); isFloat {
			f.kind = ftFloat
			break
		}
		f.kind = ftString
	}
}

func (f *fieldDescription) updateValues(value *dlit.Literal) {
	// Chose 31 so could hold each day in month
	maxNumValues := 31
	if f.kind == ftIgnore ||
		f.kind == ftUnknown ||
		f.numValues == -1 {
		return
	}
	for _, v := range f.values {
		if v.String() == value.String() {
			return
		}
	}
	if f.numValues >= maxNumValues {
		if f.kind == ftString {
			f.kind = ftIgnore
		}
		f.values = []*dlit.Literal{}
		f.numValues = -1
		return
	}
	f.values = append(f.values, value)
	f.numValues++
}

func (f *fieldDescription) updateNumBoundaries(value *dlit.Literal) {
	if f.kind == ftInt {
		valueInt, valueIsInt := value.Int()
		minInt, minIsInt := f.min.Int()
		maxInt, maxIsInt := f.max.Int()
		if !valueIsInt || !minIsInt || !maxIsInt {
			panic("Type mismatch")
		}
		f.min = dlit.MustNew(minI(minInt, valueInt))
		f.max = dlit.MustNew(maxI(maxInt, valueInt))
	} else if f.kind == ftFloat {
		valueFloat, valueIsFloat := value.Float()
		minFloat, minIsFloat := f.min.Float()
		maxFloat, maxIsFloat := f.max.Float()
		if !valueIsFloat || !minIsFloat || !maxIsFloat {
			panic("Type mismatch")
		}
		f.min = dlit.MustNew(math.Min(minFloat, valueFloat))
		f.max = dlit.MustNew(math.Max(maxFloat, valueFloat))
		f.maxDP =
			int(maxI(int64(f.maxDP), int64(numDecPlaces(value.String()))))
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
