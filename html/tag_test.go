package html

import (
	"bytes"
	"fmt"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"golang.org/x/net/html"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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
	config := &config.Config{
		WWWDir:   filepath.Join(tmpDir),
		BuildDir: "fixtures",
	}
	if err := generateTagPages(config); err != nil {
		t.Fatalf("generateTagPages(config) err: %s", err)
	}

	tagFiles, err :=
		ioutil.ReadDir(filepath.Join(config.WWWDir, "tag"))
	if err != nil {
		t.Fatalf("ioutil.ReadDir(...) err: %s", err)
	}

	tagsInfo := make(map[string]*tagInfo)
	for _, file := range tagFiles {
		if file.IsDir() {
			tagIndexFilename := filepath.Join(
				config.WWWDir,
				"tag",
				file.Name(),
				"index.html",
			)
			if tagInfo, err := getTagInfo(tagIndexFilename); err == nil {
				tagsInfo[file.Name()] = tagInfo
			}
		}
	}

	wantTagsInfo := map[string]*tagInfo{
		"bank": &tagInfo{
			"Reports for tag: bank",
			[]string{
				"reports/how-to-keep-costs-low/",
				"reports/how-to-make-a-profit/",
				"reports/how-to-make-a-loss/",
			},
		},
		"expensive": &tagInfo{
			"Reports for tag: expensive",
			[]string{
				"reports/how-to-keep-costs-low/",
			},
		},
		"fahrenheit-451": &tagInfo{
			"Reports for tag: Fahrenheit 451",
			[]string{
				"reports/how-to-keep-costs-low/",
				"reports/how-to-make-a-profit/",
				"reports/how-to-make-a-loss/",
			},
		},
		"fred-ned": &tagInfo{
			"Reports for tag: fred / ned",
			[]string{
				"reports/how-to-keep-costs-low/",
				"reports/how-to-make-a-profit/",
				"reports/how-to-make-a-loss/",
			},
		},
		"hot-in-the-city": &tagInfo{
			"Reports for tag: hot in the city",
			[]string{
				"reports/how-to-keep-costs-low/",
				"reports/how-to-make-a-profit/",
				"reports/how-to-make-a-loss/",
			},
		},
		"test": &tagInfo{
			"Reports for tag: test",
			[]string{
				"reports/how-to-make-a-profit/",
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

func getReportUrls(filename string) ([]string, error) {
	urls, err := getUrls(filename)
	if err != nil {
		return []string{}, err
	}
	reportUrls := make([]string, len(urls))
	numReportUrls := 0
	for _, url := range urls {
		if strings.HasPrefix(url, "reports/") {
			reportUrls[numReportUrls] = url
			numReportUrls++
		}
	}
	return reportUrls[:numReportUrls], nil
}

func getHref(t html.Token) string {
	for _, a := range t.Attr {
		if a.Key == "href" {
			return a.Val
		}
	}
	return ""
}

func getInnerText(z *html.Tokenizer) string {
	tt := z.Next()
	if tt == html.TextToken {
		return string(z.Text())
	}
	return ""
}

func getUrls(filename string) ([]string, error) {
	urls := make([]string, 0)
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		return urls, err
	}
	b := bytes.NewBuffer(text)
	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return urls, nil
		case html.StartTagToken:
			t := z.Token()

			if t.Data == "a" { // Is Anchor
				if url := getHref(t); len(url) > 0 {
					urls = append(urls, url)
				}
			}
		}
	}
}

func getH1(filename string) (string, error) {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	b := bytes.NewBuffer(text)
	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return "", nil
		case html.StartTagToken:
			t := z.Token()

			if t.Data == "h1" { // Is h1 header
				return getInnerText(z), nil
			}
		}
	}
}
