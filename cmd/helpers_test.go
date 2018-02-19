package cmd

import (
	"fmt"
	"sort"

	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func uniqLogEntries(entries []testhelpers.Entry) []testhelpers.Entry {
	encountered := map[string]bool{}
	result := []testhelpers.Entry{}
	for _, e := range entries {
		id := fmt.Sprintf("%d-%s", e.Level, e.Msg)
		if !encountered[id] {
			encountered[id] = true
			result = append(result, e)
		}
	}
	return result
}

func doLogEntriesMatch(got, want []testhelpers.Entry) error {
	sort.Slice(got, func(i, j int) bool {
		return got[i].Msg < got[j].Msg
	})
	sort.Slice(want, func(i, j int) bool {
		return want[i].Msg < want[j].Msg
	})
	uniqGot := uniqLogEntries(got)
	for _, eW := range want {
		found := false
		for _, eG := range uniqGot {
			if eG.Level == eW.Level && eG.Msg == eW.Msg {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("got: %v, want: %v", got, want)
		}
	}
	for _, eG := range uniqGot {
		found := false
		for _, eW := range want {
			if eG.Level == eW.Level && eG.Msg == eW.Msg {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("got: %v, want: %v", got, want)
		}
	}
	if len(got)-len(uniqGot) > 2 {
		return fmt.Errorf("too big a difference between entries got: %v\n want: %v",
			got, want)
	}
	return nil
}
