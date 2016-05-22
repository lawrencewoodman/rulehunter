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

// Package description handles the description of data in records
package description

import (
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/internal"
	"math"
)

type Description struct {
	Fields map[string]*Field
}

// Create a New Description.
func New() *Description {
	fd := map[string]*Field{}
	return &Description{fd}
}

type Field struct {
	Kind      kind
	Min       *dlit.Literal
	Max       *dlit.Literal
	MaxDP     int
	Values    []*dlit.Literal
	NumValues int
}

type kind int

const (
	UNKNOWN kind = iota
	IGNORE
	INT
	FLOAT
	STRING
)

func (fd *Field) String() string {
	return fmt.Sprintf("Kind: %s, Min: %s, Max: %s, MaxDP: %d, Values: %s",
		fd.Kind, fd.Min, fd.Max, fd.MaxDP, fd.Values)
}

func (k kind) String() string {
	switch k {
	case UNKNOWN:
		return "Unknown"
	case IGNORE:
		return "Ignore"
	case INT:
		return "Int"
	case FLOAT:
		return "Float"
	case STRING:
		return "String"
	}
	panic(fmt.Sprintf("Unsupported kind: %d", k))
}

// Analyse this record
func (d *Description) NextRecord(record map[string]*dlit.Literal) {
	if len(d.Fields) == 0 {
		for field, value := range record {
			d.Fields[field] = &Field{Kind: UNKNOWN, Min: value, Max: value}
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
	case UNKNOWN:
		fallthrough
	case INT:
		if _, isInt := value.Int(); isInt {
			f.Kind = INT
			break
		}
		fallthrough
	case FLOAT:
		if _, isFloat := value.Float(); isFloat {
			f.Kind = FLOAT
			break
		}
		f.Kind = STRING
	}
}

func (f *Field) updateValues(value *dlit.Literal) {
	// Chose 31 so could hold each day in month
	maxNumValues := 31
	if f.Kind == IGNORE || f.Kind == UNKNOWN || f.NumValues == -1 {
		return
	}
	for _, v := range f.Values {
		if v.String() == value.String() {
			return
		}
	}
	if f.NumValues >= maxNumValues {
		if f.Kind == STRING {
			f.Kind = IGNORE
		}
		f.Values = []*dlit.Literal{}
		f.NumValues = -1
		return
	}
	f.Values = append(f.Values, value)
	f.NumValues++
}

func (f *Field) updateNumBoundaries(value *dlit.Literal) {
	if f.Kind == INT {
		valueInt, valueIsInt := value.Int()
		minInt, minIsInt := f.Min.Int()
		maxInt, maxIsInt := f.Max.Int()
		if !valueIsInt || !minIsInt || !maxIsInt {
			panic("Type mismatch")
		}
		f.Min = dlit.MustNew(minI(minInt, valueInt))
		f.Max = dlit.MustNew(maxI(maxInt, valueInt))
	} else if f.Kind == FLOAT {
		valueFloat, valueIsFloat := value.Float()
		minFloat, minIsFloat := f.Min.Float()
		maxFloat, maxIsFloat := f.Max.Float()
		if !valueIsFloat || !minIsFloat || !maxIsFloat {
			panic("Type mismatch")
		}
		f.Min = dlit.MustNew(math.Min(minFloat, valueFloat))
		f.Max = dlit.MustNew(math.Max(maxFloat, valueFloat))
		f.MaxDP =
			int(maxI(int64(f.MaxDP), int64(internal.NumDecPlaces(value.String()))))
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
