package experiment

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/vlifesystems/rulehunter/config"
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
		cfg     *config.Config
		dataset ddataset.Dataset
		wantErr error
	}{
		{cfg: &config.Config{BuildDir: filepath.Join(cfgDir, "build")},
			dataset: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"group", "district", "height", "flow"},
			),
			wantErr: &os.PathError{
				"open",
				filepath.Join(cfgDir, "build", "descriptions", "aname"),
				syscall.ESRCH,
			},
		},
		{cfg: &config.Config{BuildDir: filepath.Join(cfgDir, "build")},
			dataset: dcsv.New(
				filepath.Join("fixtures", "flow_nonexistant.csv"),
				true,
				rune(','),
				[]string{"group", "district", "height", "flow"},
			),
			wantErr: &os.PathError{
				"open",
				filepath.Join("fixtures", "flow_nonexistant.csv"),
				syscall.ENOENT,
			},
		},
	}
	for _, c := range cases {
		_, err := describeDataset(c.cfg, "aname", c.dataset)
		if err == nil || err.Error() != c.wantErr.Error() {
			t.Errorf("describeDataset - gotErr: %s, wantErr: %s", err, c.wantErr)
		}
	}
}
