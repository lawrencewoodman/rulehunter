// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

const reportTpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<title>{{.Title}}</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>{{.Title}}</h1>
				Date: {{ .DateTime }} &nbsp;
				Mode: {{ .Mode }} &nbsp;
				{{if .Category}}
					Category: <a href="{{ .CategoryURL }}">{{ .Category }}</a> &nbsp;
				{{end}}
				{{if .Tags}}
					Tags:
					{{range $tag, $tagLink := .Tags}}
						<a href="{{ $tagLink }}">{{ $tag }}</a> &nbsp;
					{{end}}
				{{end}}
				<br />
				<br />
				<h2>Experiment Details</h2>
				<p>Experiment file: {{ .ExperimentFilename }}</p>
				<br />
				<table class="table table-bordered table-nonfluid">
					<tr>
						<th>Sort Order</th><th>Direction</th>
					</tr>
					{{range .SortOrder}}
						<tr>
							<td>{{ .Aggregator }}</td><td>{{ .Direction }}</td>
						</tr>
					{{end}}
				</table>

				{{if .Aggregators}}
					<table class="table table-bordered">
						<tr>
							<th>Aggregator Name</th><th>Kind</th><th>Arg</th>
						</tr>
						{{range .Aggregators}}
							<tr>
								<td>{{ .Name }}</td><td>{{ .Kind }}</td><td>{{ .Arg }}</td>
							</tr>
						{{end}}
					</table>
				{{end}}

				<h2>Data Set</h2>
				The data set contained {{ .NumRecords }} records.</br />
				<br />
				<table class="table table-bordered">
					<tr>
						<th>Field</th>
						<th>Kind</th>
						<th>Min</th>
						<th>Max</th>
						<th>MaxDP</th>
						<th>Values - ('value', freq)</th>
					</tr>
					{{range $field, $fd := .Description.Fields}}
						<tr>
							<td>{{ $field }}</td>
							<td>{{ $fd.Kind }}</td>
							{{if eq $fd.Kind.String "Number"}}
								<td>{{ $fd.Min }}</td>
								<td>{{ $fd.Max }}</td>
								<td>{{ $fd.MaxDP }}</td>
							{{else}}
								<td>N/A</td><td>N/A</td><td>N/A</td>
							{{end}}
							<td>
								{{range $value, $valueDesc := $fd.Values}}
							    ('{{ $value }}', {{ $valueDesc.Num }}) &nbsp;
								{{end}}
							</td>
						</tr>
					{{end}}
				</table>
			</div>

			<div class="container">
				<h2>Results</h2>
			</div>
			{{range .Assessments}}
				<div class="container">
					<h3>{{ .Rule }}</h3>

					<div class="pull-left aggregators">
						<table class="table table-bordered">
							<tr>
								<th>Aggregator</th>
								<th>Value</th>
								<th>Improvement</th>
							</tr>
							{{ range .Aggregators }}
							<tr>
								<td>{{ .Name }}</td>
								<td>{{ .Value }}</td>
								<td>{{ .Difference }}</td>
							</tr>
							{{ end }}
						</table>
					</div>

					{{if .Goals}}
						<div class="pull-left">
							<table class="table table-bordered">
								<tr>
									<th>Goal</th><th>Value</th>
								</tr>
								{{ range .Goals }}
								<tr>
									<td>{{ .Expr }}</td>
									<td class="goalPassed-{{.Passed}}">{{ .Passed }}</td>
								</tr>
								{{ end }}
							</table>
						</div>
					{{end}}

				</div>
			{{ end }}
		</div>

		<div id="footer" class="container">
			{{ index .Html "footer" }}
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`
