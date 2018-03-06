package html

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
)

// This tests:
//    i) That the correct reports are listed for each tag in date order
//   ii) That reports with different tags that resolve to the same
//       escaped tag are listed under a single tag
//  iii) That the shortest tag name is used if there are multiple ones that
//       resolve to the same escaped tag
func TestGenerateTagPages(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	cfg := &config.Config{
		WWWDir:   filepath.Join(tmpDir),
		BuildDir: "fixtures",
	}
	pm, err := progress.NewMonitor(tmpDir)
	if err != nil {
		t.Fatalf("NewMonitor: %s", err)
	}
	if err := generateTagPages(cfg, pm); err != nil {
		t.Fatalf("generateTagPages: %s", err)
	}

	tagFiles, err :=
		ioutil.ReadDir(filepath.Join(cfg.WWWDir, "reports", "tag"))
	if err != nil {
		t.Fatalf("ioutil.ReadDir(...) err: %s", err)
	}

	tagsInfo := make(map[string]*tagInfo)
	for _, file := range tagFiles {
		if file.IsDir() {
			tagIndexFilename := filepath.Join(
				cfg.WWWDir,
				"reports",
				"tag",
				file.Name(),
				"index.html",
			)
			if tagInfo, err := getTagInfo(tagIndexFilename); err == nil {
				tagsInfo[file.Name()] = tagInfo
			} else {
				t.Fatalf("getTagInfo: %s", err)
			}
		}
	}

	noTagIndexFilename := filepath.Join(
		cfg.WWWDir,
		"reports",
		"notag",
		"index.html",
	)
	if tagInfo, err := getTagInfo(noTagIndexFilename); err == nil {
		tagsInfo[""] = tagInfo
	} else {
		t.Fatalf("getTagInfo: %s", err)
	}

	wantTagsInfo := map[string]*tagInfo{
		"": &tagInfo{
			"Reports for tag: ",
			[]string{
				"reports/nocategory/how-to-not-contain-tags-or-cats/train/",
			},
		},
		"bank": &tagInfo{
			"Reports for tag: bank",
			[]string{
				"reports/category/groupb/how-to-keep-costs-low/train/",
				"reports/category/groupb/how-to-make-a-profit/train/",
				"reports/category/groupa/how-to-make-a-loss/train/",
			},
		},
		"expensive": &tagInfo{
			"Reports for tag: expensive",
			[]string{
				"reports/category/groupb/how-to-keep-costs-low/train/",
			},
		},
		"fahrenheit-451": &tagInfo{
			"Reports for tag: Fahrenheit 451",
			[]string{
				"reports/category/groupb/how-to-keep-costs-low/train/",
				"reports/category/groupb/how-to-make-a-profit/train/",
				"reports/category/groupa/how-to-make-a-loss/train/",
			},
		},
		"fred-ned": &tagInfo{
			"Reports for tag: fred / ned",
			[]string{
				"reports/category/groupb/how-to-keep-costs-low/train/",
				"reports/category/groupb/how-to-make-a-profit/train/",
				"reports/category/groupa/how-to-make-a-loss/train/",
			},
		},
		"hot-in-the-city": &tagInfo{
			"Reports for tag: hot in the city",
			[]string{
				"reports/category/groupb/how-to-keep-costs-low/train/",
				"reports/category/groupb/how-to-make-a-profit/train/",
				"reports/category/groupa/how-to-make-a-loss/train/",
			},
		},
		"test": &tagInfo{
			"Reports for tag: test",
			[]string{
				"reports/category/groupb/how-to-make-a-profit/train/",
			},
		},
	}

	if err := checkTagsInfoMatch(tagsInfo, wantTagsInfo); err != nil {
		t.Errorf("checkTagsInfoMatch(...) err: %s", err)
	}
}

func checkTagsInfoMatch(t1, t2 map[string]*tagInfo) error {
	if len(t1) != len(t2) {
		return fmt.Errorf("Different number of keys: %d != %d", len(t1), len(t2))
	}
	for tag, ti := range t1 {
		if err := checkTagInfoMatch(ti, t2[tag]); err != nil {
			return fmt.Errorf("Tag info: %s, doesn't match: %s", tag, err)
		}
	}
	return nil
}

func checkTagInfoMatch(ti1, ti2 *tagInfo) error {
	if ti1.h1 != ti2.h1 {
		return fmt.Errorf("h1's don't match (%s != %s)", ti1.h1, ti2.h1)
	}
	if !reflect.DeepEqual(ti1.reportUrls, ti2.reportUrls) {
		return fmt.Errorf("reportUrls's don't match (%s != %s)",
			ti1.reportUrls, ti2.reportUrls)
	}
	return nil
}

func getTagInfo(filename string) (*tagInfo, error) {
	reportUrls, err := getReportUrls(filename)
	if err != nil {
		return nil, fmt.Errorf("getReportUrls(%s) err: %s", filename, err)
	}
	h1, err := getH1(filename)
	if err != nil {
		return nil, fmt.Errorf("getH1(%s) err: %s", filename, err)
	}
	return &tagInfo{
		h1:         h1,
		reportUrls: reportUrls,
	}, nil
}

type tagInfo struct {
	h1         string
	reportUrls []string
}
