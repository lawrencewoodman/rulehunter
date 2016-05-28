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
package reducedataset

import (
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/dataset"
	"io"
)

type ReduceDataset struct {
	dataset    dataset.Dataset
	recordNum  int
	numRecords int
	err        error
}

func New(dataset dataset.Dataset, numRecords int) (dataset.Dataset, error) {
	return &ReduceDataset{
		dataset:    dataset,
		recordNum:  -1,
		numRecords: numRecords,
		err:        nil,
	}, nil
}

func (r *ReduceDataset) Clone() (dataset.Dataset, error) {
	i, err := r.dataset.Clone()
	return i, err
}

func (r *ReduceDataset) Next() bool {
	if r.Err() != nil {
		return false
	}
	if r.recordNum < r.numRecords {
		r.recordNum++
		return r.dataset.Next()
	}
	r.err = io.EOF
	return false
}

func (r *ReduceDataset) Err() error {
	if r.err == io.EOF {
		return nil
	}
	return r.dataset.Err()
}

func (r *ReduceDataset) Read() (map[string]*dlit.Literal, error) {
	record, err := r.dataset.Read()
	return record, err
}

func (r *ReduceDataset) Rewind() error {
	if r.Err() != nil {
		return r.Err()
	}
	r.recordNum = -1
	r.err = r.dataset.Rewind()
	return r.err
}

func (r *ReduceDataset) GetFieldNames() []string {
	return r.dataset.GetFieldNames()
}

func (r *ReduceDataset) Close() error {
	return r.dataset.Close()
}
