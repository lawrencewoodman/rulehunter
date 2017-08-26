/*
	rulehunter - A server to find rules in data based on user specified goals
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

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
				{{if .Category}}
					Category: <a href="{{ .CategoryURL }}">{{ .Category }}</a> &nbsp;
				{{end}}
				{{if .Tags}}
					Tags:
					{{range $tag, $catLink := .Tags}}
						<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
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
									<td>{{ .Expr }}</td><td>{{ .Passed }}</td>
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
