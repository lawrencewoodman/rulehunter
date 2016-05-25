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

const reportsTpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<meta charset="UTF-8">
		<title>Reports</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Reports</h1>

				<ul class="reports">
					{{range .Reports}}
						<li>
							<a class="title" href="{{ .Filename }}">{{ .Title }}</a><br />
							Date: {{ .DateTime }}
							Tags:
							{{range $tag, $catLink := .Tags}}
								<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
							{{end}}
						</li>
					{{end}}
				</ul>
			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`
