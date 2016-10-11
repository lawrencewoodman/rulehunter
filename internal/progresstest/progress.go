package progresstest

import (
	"errors"
	"fmt"
	"github.com/vlifesystems/rulehunter/progress"
	"math"
	"time"
)

func CheckExperimentsMatch(
	experiments1 []*progress.Experiment,
	experiments2 []*progress.Experiment,
) error {
	if len(experiments1) != len(experiments2) {
		return errors.New("Lengths of experiments don't match")
	}
	for i, e := range experiments1 {
		if err := checkExperimentMatch(e, experiments2[i]); err != nil {
			return err
		}
	}
	return nil
}

func checkExperimentMatch(e1, e2 *progress.Experiment) error {
	if e1.Title != e2.Title {
		return fmt.Errorf("Title doesn't match: %s != %s", e1.Title, e2.Title)
	}
	if e1.ExperimentFilename != e2.ExperimentFilename {
		return errors.New("ExperimentFilename doesn't match")
	}
	if e1.Msg != e2.Msg {
		return fmt.Errorf("Msg doesn't match: %s != %s", e1.Msg, e2.Msg)
	}
	if e1.Status != e2.Status {
		return errors.New("Status doesn't match")
	}
	if !timesClose(e1.Stamp, e2.Stamp, 10) {
		return errors.New("Stamp not close in time")
	}
	if len(e1.Tags) != len(e2.Tags) {
		return errors.New("Tags doesn't match")
	}
	for i, t := range e1.Tags {
		if t != e2.Tags[i] {
			return errors.New("Tags doesn't match")
		}
	}
	return nil
}

func timesClose(t1, t2 time.Time, maxSecondsDiff int) bool {
	diff := t1.Sub(t2)
	secondsDiff := math.Abs(diff.Seconds())
	return secondsDiff <= float64(maxSecondsDiff)
}
