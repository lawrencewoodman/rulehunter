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

package rule

import (
	"github.com/lawrencewoodman/ddataset"
)

// True represents a rule that always returns true
type True struct{}

func NewTrue() Rule {
	return True{}
}

func (r True) String() string {
	// TODO: Work out if should return TRUE here
	return "true()"
}

func (r True) GetInNiParts() (bool, string, string) {
	return false, "", ""
}

func (r True) IsTrue(record ddataset.Record) (bool, error) {
	return true, nil
}
