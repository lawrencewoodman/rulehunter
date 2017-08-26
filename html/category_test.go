package html

import (
	"fmt"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
	config := &config.Config{
		WWWDir:   filepath.Join(tmpDir),
		BuildDir: "fixtures",
	}
	if err := generateCategoryPages(config); err != nil {
		t.Fatalf("generateCategoryPages(config) err: %s", err)
	}

	categoryFiles, err :=
		ioutil.ReadDir(filepath.Join(config.WWWDir, "category"))
	if err != nil {
		t.Fatalf("ioutil.ReadDir(...) err: %s", err)
	}
	categoriesInfo := make(map[string]*categoryInfo)
	for _, file := range categoryFiles {
		if file.IsDir() {
			categoryIndexFilename := filepath.Join(
				config.WWWDir,
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
				"reports/how-to-make-a-loss/",
			},
		},
		"groupb": &categoryInfo{
			"Reports for category: group^^^B",
			[]string{
				"reports/how-to-keep-costs-low/",
				"reports/how-to-make-a-profit/",
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
