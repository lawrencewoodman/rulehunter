// Copyright (C) 2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package experiment

import (
	"github.com/vlifesystems/rhkit/rule"
)

type ruleTracker struct {
	rulesTracked map[string]interface{}
}

func newRuleTracker() *ruleTracker {
	return &ruleTracker{
		rulesTracked: make(map[string]interface{}, 0),
	}
}

// track adds rules to the rule tracker and returns rules not already
// tracked
func (rt *ruleTracker) track(rules []rule.Rule) []rule.Rule {
	result := []rule.Rule{}
	for _, r := range rules {
		if _, ok := rt.rulesTracked[r.String()]; !ok {
			rt.rulesTracked[r.String()] = nil
			result = append(result, r)
		}
	}
	return result
}
