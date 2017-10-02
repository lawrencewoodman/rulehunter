package html

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
)

/*************************
       Benchmarks
*************************/

func BenchmarkGenerateFront(b *testing.B) {
	cfgDir := testhelpers.BuildConfigDirs(b, true)
	defer os.RemoveAll(cfgDir)

	cfg := &config.Config{
		ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
		WWWDir:            filepath.Join(cfgDir, "www"),
		BuildDir:          filepath.Join(cfgDir, "build"),
		MaxNumRecords:     100,
		MaxNumProcesses:   4,
		MaxNumReportRules: 100,
	}
	htmlCmds := make(chan cmd.Cmd, 100)
	defer close(htmlCmds)
	cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.Run()
	pm, err := progress.NewMonitor(
		filepath.Join(cfg.BuildDir, "progress"),
		htmlCmds,
	)
	if err != nil {
		b.Fatalf("progress.NewMonitor: %s", err)
	}

	for i := 0; i < 100; i++ {
		testhelpers.CopyFile(
			b,
			filepath.Join("fixtures", "reports", "bank-profit.json"),
			filepath.Join(cfgDir, "build", "reports"),
			fmt.Sprintf("bank-profit-%02d.json", i),
		)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := generateFront(cfg, pm); err != nil {
			b.Fatalf("generateFront: %s", err)
		}
	}

}
