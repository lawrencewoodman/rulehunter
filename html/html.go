/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */

package html

import (
	"fmt"
	"github.com/lawrencewoodman/rulehunter"
	"github.com/lawrencewoodman/rulehuntersrv/config"
	"github.com/lawrencewoodman/rulehuntersrv/progress"
	"github.com/lawrencewoodman/rulehuntersrv/report"
	"hash/crc32"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func GenerateReports(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
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
      Categories:
      {{range $category, $catLink := .Categories}}
		    <a href="{{ $catLink }}">{{ $category }}</a> &nbsp;
      {{end}}<br />
			<br />
		{{end}}
	</body>
</html>`

	type TplReport struct {
		Title      string
		Categories map[string]string
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
				makeCategoryLinks(report.Categories),
				report.Stamp.Format(time.RFC822),
				reportFilename,
			}
		}
		i++
	}
	tplData := TplData{tplReports}

	if err := generateCategoryPages(config); err != nil {
		return err
	}

	if err := generateProgressPage(config, progressMonitor); err != nil {
		return err
	}

	outputFilename := filepath.Join(config.WWWDir, "reports", "index.html")
	return writeTemplate(outputFilename, tpl, tplData)
}

func generateProgressPage(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="refresh" content="4" >
		<title>Progress</title>
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
		<h1>Progress</h1>

		{{range .Experiments}}
			<strong>{{ .Title }}</strong><br />
			Date: {{ .Stamp }}
      Categories:
      {{range $category, $catLink := .Categories}}
		    <a href="{{ $catLink }}">{{ $category }}</a> &nbsp;
      {{end}}<br />
			Experiment filename: {{ .Filename }}<br />
		  Status: {{ .Status }} &nbsp; Message: {{ .Msg }}<br />
			<br />
		{{end}}
	</body>
</html>`

	type TplExperiment struct {
		Title      string
		Categories map[string]string
		Stamp      string
		Filename   string
		Status     string
		Msg        string
	}

	type TplData struct {
		Experiments []*TplExperiment
	}

	experiments, err := progressMonitor.GetExperiments()
	if err != nil {
		return err
	}

	tplExperiments := make([]*TplExperiment, len(experiments))

	for i, experiment := range experiments {
		tplExperiments[i] = &TplExperiment{
			experiment.Title,
			makeCategoryLinks(experiment.Categories),
			experiment.Stamp.Format(time.RFC822),
			experiment.ExperimentFilename,
			experiment.Status.String(),
			experiment.Msg,
		}
	}
	tplData := TplData{tplExperiments}

	outputFilename := filepath.Join(config.WWWDir, "progress", "index.html")
	return writeTemplate(outputFilename, tpl, tplData)
}

func generateCategoryPage(
	config *config.Config,
	categoryName string,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Reports for category: {{ .Category }}</title>
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
		<h1>Reports for category: {{ .Category }}</h1>

		{{range .Reports}}
			<a class="title" href="../../{{ .Filename }}">{{ .Title }}</a><br />
			Date: {{ .Stamp }}
      Categories:
      {{range $category, $catLink := .Categories}}
		    <a href="../../{{ $catLink }}">{{ $category }}</a> &nbsp;
      {{end}}<br />
			<br />
		{{end}}
	</body>
</html>`

	type TplReport struct {
		Title      string
		Categories map[string]string
		Stamp      string
		Filename   string
	}

	type TplData struct {
		Category string
		Reports  []*TplReport
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
			if inStrings(categoryName, report.Categories) {
				reportFilename := makeReportFilename(report.Stamp, report.Title)
				tplReports[i] = &TplReport{
					report.Title,
					makeCategoryLinks(report.Categories),
					report.Stamp.Format(time.RFC822),
					filepath.Join(reportFilename),
				}
				i++
			}
		}
	}
	tplReports = tplReports[:i]
	tplData := TplData{categoryName, tplReports}
	fullCategoryDir := filepath.Join(
		config.WWWDir,
		"reports",
		"category",
		escapeString(categoryName),
	)

	if err := os.MkdirAll(fullCategoryDir, 0740); err != nil {
		return err
	}
	outputFilename := filepath.Join(fullCategoryDir, "index.html")
	return writeTemplate(outputFilename, tpl, tplData)
}

func inStrings(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func generateCategoryPages(config *config.Config) error {
	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	categoriesSeen := make(map[string]bool)
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJson(config, file.Name())
			if err != nil {
				return err
			}
			for _, category := range report.Categories {
				if _, ok := categoriesSeen[category]; !ok {
					if err := generateCategoryPage(config, category); err != nil {
						return err
					}
					categoriesSeen[category] = true
				}
			}
		}
	}
	return nil
}

func makeCategoryLinks(categories []string) map[string]string {
	catLinks := make(map[string]string, len(categories))
	for _, category := range categories {
		catLinks[category] = makeCategoryLink(category)
	}
	return catLinks
}

func makeCategoryLink(category string) string {
	return fmt.Sprintf(
		"category/%s/index.html",
		escapeString(category),
	)
}

var nonAlphaNumRegexp = regexp.MustCompile("[^[:alnum:]]")

func escapeString(s string) string {
	crc32 := strconv.FormatUint(
		uint64(crc32.Checksum([]byte(s), crc32.MakeTable(crc32.IEEE))),
		36,
	)
	newS := nonAlphaNumRegexp.ReplaceAllString(s, "")
	return fmt.Sprintf("%s_%s", newS, crc32)
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
					<td class="last-column">
						{{range $category, $catLink := .Categories}}
							<a href="{{ $catLink }}">{{ $category }}</a> &nbsp;
						{{end}}<br />
					</td>
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

	type TplData struct {
		Title              string
		Categories         map[string]string
		Stamp              string
		ExperimentFilename string
		NumRecords         int64
		SortOrder          []rulehunter.SortField
		Assessments        []*report.Assessment
	}

	categoryLinks := makeCategoryLinks(_report.Categories)

	tplData := TplData{
		_report.Title,
		categoryLinks,
		_report.Stamp.Format(time.RFC822),
		_report.ExperimentFilename,
		_report.NumRecords,
		_report.SortOrder,
		_report.Assessments,
	}

	fullReportFilename := filepath.Join(config.WWWDir, "reports", reportFilename)
	if err := writeTemplate(fullReportFilename, tpl, tplData); err != nil {
		return err
	}
	return nil
}

func writeTemplate(
	filename string,
	tpl string,
	tplData interface{},
) error {
	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := t.Execute(f, tplData); err != nil {
		return err
	}
	return nil
}

func makeReportFilename(stamp time.Time, title string) string {
	timeSeconds := strconv.FormatInt(stamp.Unix(), 36)
	escapedTitle := escapeString(title)
	return fmt.Sprintf("%s_%s.html", escapedTitle, timeSeconds)
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
