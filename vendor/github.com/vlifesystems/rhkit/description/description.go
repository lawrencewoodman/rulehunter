// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// Package description handles describing a Dataset
package description

import (
	"fmt"
	"sort"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/internal"
)

// Description describes a Dataset
type Description struct {
	Fields map[string]*Field `json:"fields"`
}

// Value describes a value in a field
type Value struct {
	Value *dlit.Literal
	Num   int
}

type InvalidFieldError string

func (e InvalidFieldError) Error() string {
	return "invalid field: " + string(e)
}

// DescribeDataset analyses a Dataset and returns a Description of it
func DescribeDataset(dataset ddataset.Dataset) (*Description, error) {
	if err := checkFieldsValid(dataset.Fields()); err != nil {
		return nil, err
	}
	desc := newDescription()
	conn, err := dataset.Open()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	for conn.Next() {
		record := conn.Read()
		desc.nextRecord(record)
	}

	return desc, conn.Err()
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

// CheckEqual checks if two Descriptions are equal
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

func (d *Description) FieldNames() []string {
	names := make([]string, len(d.Fields))
	i := 0
	for n := range d.Fields {
		names[i] = n
		i++
	}
	return names
}

// newDescription creates a new Description
func newDescription() *Description {
	fd := map[string]*Field{}
	return &Description{fd}
}

// nextRecord updates the description after analysing the supplied record
func (d *Description) nextRecord(record ddataset.Record) {
	if len(d.Fields) == 0 {
		for field, value := range record {
			d.Fields[field] = &Field{
				Kind:   Unknown,
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

func checkFieldsValid(fields []string) error {
	for _, field := range fields {
		if !internal.IsIdentifierValid(field) {
			return InvalidFieldError(field)
		}
	}
	return nil
}
