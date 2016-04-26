/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */

package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type indexReportDesc struct {
	Title      string
	Stamp      time.Time
	Categories []string
	Filename   string
}

type indexFile struct {
	Reports []*indexReportDesc
}

func loadReportsIndex(buildDir string) ([]*indexReportDesc, error) {
	var index indexFile
	filename := filepath.Join(buildDir, "index.json")

	f, err := os.Open(filename)
	if err != nil {
		return []*indexReportDesc{}, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err = dec.Decode(&index); err != nil {
		return []*indexReportDesc{}, err
	}
	return index.Reports, nil
}

func addReportToIndex(
	buildDir string,
	reportFilename string,
	title string,
	tags []string,
) error {
	indexFilename := filepath.Join(buildDir, "index.json")
	reportDescs, err := loadReportsIndex(buildDir)
	if err != nil {
		return err
	}
	newReportDesc := &indexReportDesc{title, time.Now(), tags, reportFilename}
	reportDescs = append([]*indexReportDesc{newReportDesc}, reportDescs...)
	index := indexFile{reportDescs}
	json, err := json.Marshal(index)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(indexFilename, json, 0640); err != nil {
		return err
	}

	return nil
}

func writeIndexHTML(buildDir, reportsDir string) error {
	reportDescs, err := loadReportsIndex(buildDir)
	if err != nil {
		return err
	}

	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Reports</title>
	</head>

	<style>
		a.title {
			color: black;
			font-size: 120%;
			text-decoration: none;
			font-weight: bold;
		}
		a.title:visited {
			color: black;
		}
		a.title:hover {
			text-decoration: underline;
		}
	</style>

	<body>
		<h1>Reports</h1>

		{{range .Reports}}
			<a class="title" href="{{ .Filename }}">{{ .Title }}</a><br />
			Date: {{ .Stamp }}
      Categories: {{range .Categories}} {{ . }} {{end}}<br />
			<br />
		{{end}}
	</body>
</html>`

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return err
	}
	type TplReport struct {
		Title      string
		Categories []string
		Stamp      string
		Filename   string
	}

	type TplData struct {
		Reports []*TplReport
	}

	tplReports := make([]*TplReport, len(reportDescs))
	for i, reportDesc := range reportDescs {
		tplReports[i] = &TplReport{
			reportDesc.Title,
			reportDesc.Categories,
			reportDesc.Stamp.Format(time.RFC822),
			reportDesc.Filename,
		}
	}
	tplData := TplData{tplReports}

	outputFilename := filepath.Join(reportsDir, "index.html")
	f, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, tplData)
}
