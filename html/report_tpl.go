/*
	rulehuntersrv - A server to find rules in data based on user specified goals
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>

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

				<h2>Config</h2>
				<table class="neat-table">
					<tr class="title">
						<th> </th>
						<th class="last-column"> </th>
					</tr>
					<tr>
						<td>Tags</td>
						<td class="last-column">
							{{range $tag, $catLink := .Tags}}
								<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
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

				<table class="neat-table">
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

			<div class="container">
				<h2>Results</h2>
			</div>
			{{range .Assessments}}
				<div class="container">
					<h3>{{ .Rule }}</h3>

					<div class="pull-left aggregators">
						<table class="neat-table">
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

					<div class="pull-left">
						<table class="neat-table">
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
