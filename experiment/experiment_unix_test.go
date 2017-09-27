// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package experiment

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestDescribeDataset_errors(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, false)
	defer os.RemoveAll(cfgDir)
	cases := []struct {
		cfg        *config.Config
		experiment *Experiment
		dataset    ddataset.Dataset
		wantErr    error
	}{
		{cfg: &config.Config{BuildDir: filepath.Join(cfgDir, "build")},
			experiment: &Experiment{
				Title:    "What would indicate good flow?",
				Category: "testing",
				Dataset: dcsv.New(
					filepath.Join("fixtures", "flow.csv"),
					true,
					rune(','),
					[]string{"group", "district", "height", "flow"},
				),
			},
			wantErr: &os.PathError{
				"open",
				filepath.Join(
					cfgDir,
					"build",
					"descriptions",
					internal.MakeBuildFilename("testing", "What would indicate good flow?"),
				),
				syscall.ENOENT,
			},
		},
		{cfg: &config.Config{BuildDir: filepath.Join(cfgDir, "build")},
			experiment: &Experiment{
				Dataset: dcsv.New(
					filepath.Join("fixtures", "flow_nonexistant.csv"),
					true,
					rune(','),
					[]string{"group", "district", "height", "flow"},
				),
			},
			wantErr: &os.PathError{
				"open",
				filepath.Join("fixtures", "flow_nonexistant.csv"),
				syscall.ENOENT,
			},
		},
	}
	for _, c := range cases {
		_, err := c.experiment.describeDataset(c.cfg)
		if err == nil || err.Error() != c.wantErr.Error() {
			t.Errorf("describeDataset - gotErr: %s, wantErr: %s", err, c.wantErr)
		}
	}
}
