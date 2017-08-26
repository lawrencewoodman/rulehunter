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
		return fmt.Errorf("Lengths of experiments don't match: %d != %d",
			len(experiments1), len(experiments2))
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
	if e1.Filename != e2.Filename {
		return fmt.Errorf("Filename doesn't match: %s != %s",
			e1.Filename, e2.Filename)
	}
	if e1.Status.Msg != e2.Status.Msg {
		return fmt.Errorf("Status.Msg doesn't match: %s != %s",
			e1.Status.Msg, e2.Status.Msg)
	}
	if e1.Status.Percent != e2.Status.Percent {
		return errors.New("Status.Percent doesn't match")
	}
	if e1.Status.State != e2.Status.State {
		return errors.New("Status.State doesn't match")
	}
	if !timesClose(e1.Status.Stamp, e2.Status.Stamp, 10) {
		return errors.New("Status.Stamp not close in time")
	}
	if len(e1.Tags) != len(e2.Tags) {
		return errors.New("Tags doesn't match")
	}
	for i, t := range e1.Tags {
		if t != e2.Tags[i] {
			return errors.New("Tags doesn't match")
		}
	}
	if e1.Category != e2.Category {
		return errors.New("Categories don't match")
	}
	return nil
}

func timesClose(t1, t2 time.Time, maxSecondsDiff int) bool {
	diff := t1.Sub(t2)
	secondsDiff := math.Abs(diff.Seconds())
	return secondsDiff <= float64(maxSecondsDiff)
}
