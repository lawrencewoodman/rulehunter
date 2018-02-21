package experiment

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestShouldProcess(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	cases := []struct {
		file       fileinfo.FileInfo
		when       string
		isFinished bool
		pmStamp    time.Time
		want       bool
	}{
		{file: testhelpers.NewFileInfo("bank-divorced.json", time.Now()),
			when:       "!hasRun",
			isFinished: true,
			pmStamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-04T14:53:00.570347516+01:00"),
			want: true,
		},
		{file: testhelpers.NewFileInfo("bank-divorced.json",
			testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-04T14:53:00.570347516+01:00")),
			when:       "!hasRun",
			isFinished: true,
			pmStamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-04T14:53:00.570347516+01:00"),
			want: false,
		},
		{file: testhelpers.NewFileInfo("bank-tiny.json", time.Now()),
			when:       "!hasRun",
			isFinished: true,
			pmStamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-05T09:37:58.220312223+01:00"),
			want: true,
		},
		{file: testhelpers.NewFileInfo("bank-tiny.json",
			testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-05T09:37:58.220312223+01:00")),
			when:       "!hasRun",
			isFinished: true,
			pmStamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-05T09:37:58.220312223+01:00"),
			want: false,
		},
		{file: testhelpers.NewFileInfo("bank-full-divorced.json", time.Now()),
			when:       "!hasRun",
			isFinished: false,
			pmStamp:    time.Now(),
			want:       true,
		},
		{file: testhelpers.NewFileInfo("bank-full-divorced.json", time.Now()),
			when:       "!hasRun",
			isFinished: false,
			pmStamp:    time.Now(),
			want:       true,
		},
	}
	cfg := &config.Config{
		MaxNumRecords: -1,
		BuildDir:      filepath.Join(tmpDir, "build"),
	}

	fields := []string{"name", "balance", "num_cards", "marital_status",
		"tertiary_educated", "success",
	}
	testhelpers.CopyFile(
		t,
		filepath.Join("..", "progress", "fixtures", "progress.json"),
		tmpDir,
	)

	for i, c := range cases {
		desc := &modeDesc{
			Dataset: &datasetDesc{
				CSV: &csvDesc{
					Filename:  filepath.Join("fixtures", "debt.csv"),
					HasHeader: false,
					Separator: ",",
				},
			},
			When: c.when,
		}
		m, err := newMode("train", cfg, fields, desc)
		if err != nil {
			t.Fatalf("newMode: %s", err)
		}
		got, err := m.ShouldProcess(c.file, c.isFinished, c.pmStamp)
		if err != nil {
			t.Errorf("(%d) shouldProcess: %s", i, err)
			continue
		}
		if got != c.want {
			t.Errorf("(%d) shouldProcess, got: %t, want: %t", i, got, c.want)
		}
	}
}
