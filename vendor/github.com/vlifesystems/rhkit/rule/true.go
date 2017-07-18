// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

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

func (r True) IsTrue(record ddataset.Record) (bool, error) {
	return true, nil
}

func (r True) Fields() []string {
	return []string{}
}
