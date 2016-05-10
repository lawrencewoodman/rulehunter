/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */

package html

import (
	"bytes"
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

func Generate(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	if err := generateHomePage(config, progressMonitor); err != nil {
		return err
	}
	if err := generateReports(config, progressMonitor); err != nil {
		return err
	}
	if err := generateCategoryPages(config); err != nil {
		return err
	}
	return generateProgressPage(config, progressMonitor)
}

func generateHomePage(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<meta charset="UTF-8">
		<title>Rulehunter</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Rulehunter</h1>
			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`

	type TplData struct {
		Html map[string]template.HTML
	}

	tplData := TplData{
		makeHtml("home"),
	}

	outputFilename := filepath.Join(config.WWWDir, "index.html")
	return writeTemplate(outputFilename, tpl, tplData)
}

func generateReports(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
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
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
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
			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
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
		Html    map[string]template.HTML
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
	tplData := TplData{
		tplReports,
		makeHtml("reports"),
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
		{{ index .Html "head" }}
		<meta http-equiv="refresh" content="4">
    <title>Progress</title>
  </head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
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
			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
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
		Html        map[string]template.HTML
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
	tplData := TplData{tplExperiments, makeHtml("progress")}

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
		{{ index .Html "head" }}
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
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Reports for category: {{ .Category }}</h1>

				{{range .Reports}}
					<a class="title" href="{{ .Filename }}">{{ .Title }}</a><br />
					Date: {{ .Stamp }}
					Categories:
					{{range $category, $catLink := .Categories}}
						<a href="{{ $catLink }}">{{ $category }}</a> &nbsp;
					{{end}}<br />
					<br />
				{{end}}
			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
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
		Html     map[string]template.HTML
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
					fmt.Sprintf("/reports/%s", reportFilename),
				}
				i++
			}
		}
	}
	tplReports = tplReports[:i]
	tplData := TplData{categoryName, tplReports, makeHtml("category")}
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
		"/reports/category/%s/",
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
		{{ index .Html "head" }}
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

		div.rule-assessment {
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
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
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
			</div>

			<div class="container">
				<h2>Results</h2>
			</div>
			{{range .Assessments}}
				<div class="container rule-assessment">
					<h3>{{ .Rule }}</h3>

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
							<tr class="title">
								<th>Goal</th><th class="last-column">Value</th>
							</tr>
							{{ range .Goals }}
							<tr>
								<td>{{ .Expr }}</td><td class="last-column">{{ .Passed }}</td>
							</tr>
							{{ end }}
						</table>
					</div>

				</div>
			{{ end }}
		</div>

		{{ index .Html "bootstrapJS" }}
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
		Html               map[string]template.HTML
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
		makeHtml("reports"),
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

const htmlHead = `
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <!-- The above 3 meta tags *must* come first in the head; any other head content must come *after* these tags -->

    <link href="/css/bootstrap.min.css" rel="stylesheet">
    <link href="/css/sitestyle.css" rel="stylesheet">

    <!-- HTML5 shim and Respond.js for IE8 support of HTML5 elements and media queries -->
    <!-- WARNING: Respond.js doesn't work if you view the page via file:// -->
    <!--[if lt IE 9]>
      <script src="https://oss.maxcdn.com/html5shiv/3.7.2/html5shiv.min.js"></script>
      <script src="https://oss.maxcdn.com/respond/1.4.2/respond.min.js"></script>
    <![endif]-->`

const htmlBootstrapJS = `
		<!-- jQuery (necessary for Bootstrap's JavaScript plugins) -->
			<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
			<!-- Include all compiled plugins (below), or include individual files as needed -->
			<script src="/js/bootstrap.min.js"></script>`

func makeHtmlNav(menuItem string) template.HTML {
	const tpl = `
    <nav class="navbar navbar-inverse navbar-fixed-top">
      <div class="container">
        <div class="navbar-header">
          <button type="button" class="navbar-toggle collapsed"
                  data-toggle="collapse" data-target="#navbar"
								  aria-expanded="false" aria-controls="navbar">
            <span class="sr-only">Toggle navigation</span>
            <span class="icon-bar"></span>
            <span class="icon-bar"></span>
            <span class="icon-bar"></span>
          </button>
          <a class="navbar-brand" href="/">RuleHunter</a>
        </div>
        <div id="navbar" class="collapse navbar-collapse">
          <ul class="nav navbar-nav">
						{{if eq .MenuItem "home"}}
							<li class="active"><a href="/">Home</a></li>
						{{else}}
							<li><a href="/">Home</a></li>
						{{end}}

						{{if eq .MenuItem "reports"}}
							<li class="active"><a href="/reports/">Reports</a></li>
						{{else}}
							<li><a href="/reports/">Reports</a></li>
						{{end}}

						{{if eq .MenuItem "category"}}
							<li class="active"><a href=".">Category</a></li>
						{{end}}

						{{if eq .MenuItem "progress"}}
							<li class="active"><a href="/progress/">Progress</a></li>
						{{else}}
							<li><a href="/progress/">Progress</a></li>
						{{end}}
          </ul>
        </div><!--/.nav-collapse -->
      </div>
    </nav>`

	var doc bytes.Buffer
	validMenuItems := []string{
		"home",
		"reports",
		"category",
		"progress",
	}

	foundValidItem := false
	for _, validMenuItem := range validMenuItems {
		if validMenuItem == menuItem {
			foundValidItem = true
		}
	}
	if !foundValidItem {
		panic(fmt.Sprintf("menuItem not valid: %s", menuItem))
	}

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create nav html: %s",
			menuItem, err))
	}

	tplData := struct{ MenuItem string }{menuItem}

	if err := t.Execute(&doc, tplData); err != nil {
		panic(fmt.Sprintf("Couldn't create nav html: %s",
			menuItem, err))
	}
	return template.HTML(doc.String())
}

func makeHtml(menuItem string) map[string]template.HTML {
	r := make(map[string]template.HTML)
	r["head"] = template.HTML(htmlHead)
	r["nav"] = makeHtmlNav(menuItem)
	r["bootstrapJS"] = template.HTML(htmlBootstrapJS)
	return r
}
