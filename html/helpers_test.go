package html

import (
	"bytes"
	"golang.org/x/net/html"
	"io/ioutil"
	"strings"
)

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
