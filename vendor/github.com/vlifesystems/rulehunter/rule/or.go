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

package rule

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
)

// Or represents a rule determening if ruleA OR ruleB
type Or struct {
	ruleA Rule
	ruleB Rule
}

func NewOr(ruleA Rule, ruleB Rule) Rule {
	return &Or{ruleA: ruleA, ruleB: ruleB}
}

func (r *Or) String() string {
	// TODO: Consider making this OR rather than ||
	return fmt.Sprintf("%s || %s", r.ruleA, r.ruleB)
}

func (r *Or) GetInNiParts() (bool, string, string) {
	return false, "", ""
}

func (r *Or) IsTrue(record ddataset.Record) (bool, error) {
	lh, err := r.ruleA.IsTrue(record)
	if err != nil {
		return false, InvalidRuleError{Rule: r}
	}
	rh, err := r.ruleB.IsTrue(record)
	if err != nil {
		return false, InvalidRuleError{Rule: r}
	}
	return lh || rh, nil
}
