/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */

package html

import (
	"fmt"
	"github.com/lawrencewoodman/rulehunter"
	"github.com/lawrencewoodman/rulehuntersrv/config"
	"github.com/lawrencewoodman/rulehuntersrv/report"
	"hash/crc64"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func GenerateReports(config *config.Config) error {
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

	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	numReportFiles := countFiles(reportFiles)
	tplReports := make([]*TplReport, numReportFiles)

	i := 0
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJson(config, file.Name())
			if err != nil {
				return err
			}
			reportFilename := makeReportFilename(report.Stamp, report.Title)
			if err = generateReport(report, reportFilename, config); err != nil {
				return err
			}
			tplReports[i] = &TplReport{
				report.Title,
				report.Categories,
				report.Stamp.Format(time.RFC822),
				filepath.Join(reportFilename),
			}
		}
		i++
	}
	tplData := TplData{tplReports}

	outputFilename := filepath.Join(config.WWWDir, "reports", "index.html")
	f, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, tplData)
}

func generateReport(
	_report *report.Report,
	reportFilename string,
	config *config.Config,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>

	<style>
		table {
			border-collapse: collapse;
		}

		table th {
			text-align: left;
			padding-left: 1em;
			padding-right: 2em;
			border-collapse: collapse;
			border-right: 1px solid black;
			border-bottom: 1px solid black;
		}

		table th.last-column {
			border-right: 0;
		}
		table td {
			border-collapse: collapse;
			border-right: 1px solid black;
			padding-left: 1em;
			padding-right: 1em;
		  padding-top: 0.1em;
		  padding-bottom: 0.1em;
		}
		table td.last-column {
			border-right: 0;
		}
		table tr.title td {
			border-bottom: 1px solid black;
		}
		div {
			margin-bottom: 2em;
		}
		div.aggregators {
			float: left;
		  clear: left;
			margin-right: 3em;
		}
		div.goals {
			float: left;
		  clear: right;
		}

		div.config table {
			margin-bottom: 2em;
		}
		div.config table th {
			height: 1em;
		}
	</style>

	<body>
		<h1>{{.Title}}</h1>

		<div class="config">
			<h2>Config</h2>
			<table>
				<tr class="title">
					<th> </th>
					<th class="last-column"> </th>
				</tr>
				<tr>
					<td>Categories</td>
					<td class="last-column">{{range .Categories}} {{ . }} {{end}}</td>
				</tr>
				<tr>
					<td>Number of records</td>
					<td class="last-column">{{.NumRecords}}</td>
				</tr>
				<tr>
					<td>Experiment file</td>
					<td class="last-column">{{.ExperimentFilename}}</td>
				</tr>
			</table>

			<table>
				<tr class="title">
					<th>Sort Order</th><th class="last-column">Direction</th>
				</tr>
				{{range .SortOrder}}
					<tr>
						<td>{{ .Field }}</td><td class="last-column">{{ .Direction }}</td>
					</tr>
				{{end}}
			</table>
		</div>

		<h2>Results</h2>
		{{range .Assessments}}
			<h3 style="clear: both;">{{ .Rule }}</h3>


			<div class="aggregators">
				<table>
					<tr class="title">
						<th>Aggregator</th>
            <th>Value</th>
						<th class="last-column">Improvement</th>
					</tr>
					{{ range .Aggregators }}
					<tr>
						<td>{{ .Name }}</td>
						<td>{{ .Value }}</td>
						<td class="last-column">{{ .Difference }}</td>
					</tr>
					{{ end }}
				</table>
			</div>

			<div class="goals">
				<table>
					<tr class="title"><th>Goal</th><th class="last-column">Value</th></tr>
					{{ range .Goals }}
					<tr>
						<td>{{ .Expr }}</td><td class="last-column">{{ .Passed }}</td>
					</tr>
					{{ end }}
				</table>
			</div>
		{{ end }}
	</body>
</html>`

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return err
	}

	type TplData struct {
		Title              string
		Categories         []string
		Stamp              string
		ExperimentFilename string
		NumRecords         int64
		SortOrder          []rulehunter.SortField
		Assessments        []*report.Assessment
	}

	tplData := TplData{
		_report.Title,
		_report.Categories,
		_report.Stamp.Format(time.RFC822),
		_report.ExperimentFilename,
		_report.NumRecords,
		_report.SortOrder,
		_report.Assessments,
	}

	fullReportFilename := filepath.Join(config.WWWDir, "reports", reportFilename)
	f, err := os.Create(fullReportFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, tplData)
}

func makeReportFilename(stamp time.Time, title string) string {
	timeSeconds := strconv.FormatInt(stamp.Unix(), 36)
	crc64 := strconv.FormatUint(
		crc64.Checksum([]byte(title), crc64.MakeTable(crc64.ISO)),
		36,
	)
	return fmt.Sprintf("report_%s_%s.html", timeSeconds, crc64)
}

func countFiles(files []os.FileInfo) int {
	numFiles := 0
	for _, file := range files {
		if !file.IsDir() {
			numFiles++
		}
	}
	return numFiles
}
