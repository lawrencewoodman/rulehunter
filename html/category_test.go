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
//    i) That the correct reports are listed for each category in date order
//   ii) That reports with different categories that resolve to the same
//       escaped category are listed under a single category
//  iii) That the shortest category name is used if there are multiple ones that
//       resolve to the same escaped category
func TestGenerateCategoryPages(t *testing.T) {
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
	if err := generateCategoryPages(cfg, pm); err != nil {
		t.Fatalf("generateCategoryPages: %s", err)
	}

	categoryFiles, err :=
		ioutil.ReadDir(filepath.Join(cfg.WWWDir, "reports", "category"))
	if err != nil {
		t.Fatalf("ioutil.ReadDir(...) err: %s", err)
	}
	categoriesInfo := make(map[string]*categoryInfo)
	for _, file := range categoryFiles {
		if file.IsDir() {
			categoryIndexFilename := filepath.Join(
				cfg.WWWDir,
				"reports",
				"category",
				file.Name(),
				"index.html",
			)
			categoryInfo, err := getCategoryInfo(categoryIndexFilename)
			if err == nil {
				categoriesInfo[file.Name()] = categoryInfo
			}
		}
	}

	wantCategoriesInfo := map[string]*categoryInfo{
		"groupa": &categoryInfo{
			"Reports for category: groupA",
			[]string{
				"reports/category/groupa/how-to-make-a-loss/train/",
			},
		},
		"groupb": &categoryInfo{
			"Reports for category: group^^^B",
			[]string{
				"reports/category/groupb/how-to-keep-costs-low/train/",
				"reports/category/groupb/how-to-make-a-profit/train/",
			},
		},
	}

	err = checkCategoriesInfoMatch(categoriesInfo, wantCategoriesInfo)
	if err != nil {
		t.Errorf("checkCategoriesInfoMatch: %s", err)
	}
}

func checkCategoriesInfoMatch(c1, c2 map[string]*categoryInfo) error {
	if len(c1) != len(c2) {
		return fmt.Errorf("Different number of keys: %d != %d", len(c1), len(c2))
	}
	for category, ci := range c1 {
		if err := checkCategoryInfoMatch(ci, c2[category]); err != nil {
			return fmt.Errorf("Category info: %s, doesn't match: %s", category, err)
		}
	}
	return nil
}

func checkCategoryInfoMatch(ci1, ci2 *categoryInfo) error {
	if ci1.h1 != ci2.h1 {
		return fmt.Errorf("h1's don't match (%s != %s)", ci1.h1, ci2.h1)
	}
	if !reflect.DeepEqual(ci1.reportUrls, ci2.reportUrls) {
		return fmt.Errorf("reportUrls's don't match (%s != %s)",
			ci1.reportUrls, ci2.reportUrls)
	}
	return nil
}

func getCategoryInfo(filename string) (*categoryInfo, error) {
	reportUrls, err := getReportUrls(filename)
	if err != nil {
		return nil, fmt.Errorf("getReportUrls(%s) err: %s", filename, err)
	}
	h1, err := getH1(filename)
	if err != nil {
		return nil, fmt.Errorf("getH1(%s) err: %s", filename, err)
	}
	return &categoryInfo{
		h1:         h1,
		reportUrls: reportUrls,
	}, nil
}

type categoryInfo struct {
	h1         string
	reportUrls []string
}
