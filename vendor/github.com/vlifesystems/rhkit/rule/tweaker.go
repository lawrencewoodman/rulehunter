// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import (
	"github.com/vlifesystems/rhkit/description"
)

func Tweak(
	stage int,
	rules []Rule,
	inputDescription *description.Description,
) []Rule {
	newRules := make([]Rule, 0)
	for _, r := range rules {
		switch x := r.(type) {
		case Tweaker:
			rules := x.Tweak(inputDescription, stage)
			newRules = append(newRules, rules...)
		}
	}
	newRules = append(newRules, NewTrue())
	return Uniq(newRules)
}
