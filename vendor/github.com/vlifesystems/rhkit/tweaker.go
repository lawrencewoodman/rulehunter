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
	"github.com/vlifesystems/rhkit/rule"
)

func TweakRules(
	stage int,
	sortedRules []rule.Rule,
	inputDescription *Description,
) []rule.Rule {
	newRules := make([]rule.Rule, 0)
	for _, r := range sortedRules {
		switch x := r.(type) {
		case rule.Tweaker:
			field := r.GetFields()[0]
			min := inputDescription.fields[field].min
			max := inputDescription.fields[field].max
			maxDP := inputDescription.fields[field].maxDP
			rules := x.Tweak(min, max, maxDP, stage)
			newRules = append(newRules, rules...)
		}
	}
	newRules = append(newRules, rule.NewTrue())
	return rule.Uniq(newRules)
}
