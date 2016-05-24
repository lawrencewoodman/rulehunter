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

const progressTpl = `
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

				<ul class="reports-progress">
				{{range .Experiments}}
					<li>
						<table class="table table-bordered">
						  <tr>
								<th class="report-progress-th">Date</th>
								<td>{{ .Stamp }}</td>
							</tr>
							{{if .Title}}
								<tr><th>Title</th><td>{{ .Title }}</td></tr>
							{{end}}
							{{if .Tags}}
								<tr>
									<th>Tags</th>
									<td>
										{{range $tag, $catLink := .Tags}}
											<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
										{{end}}
									</td>
								</tr>
							{{end}}
							<tr><th>Experiment filename</th><td>{{ .Filename }}</td></tr>
							<tr><th>Message</th><td>{{ .Msg }}</td></tr>
							<tr>
								<th>Status</th>
								<td class="status-{{ .Status }}">{{ .Status }}</td>
							</tr>
						</table>
					</li>
				{{end}}
				</ul>
			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`
