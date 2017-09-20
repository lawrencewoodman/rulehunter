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

const frontTpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<title>Rulehunter</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">

				<div class="row">
				  <div class="col-md-12">
						<div class="page-header">
							<h1>Rulehunter<br />
								<small>Find simple rules in your data to meet your goals</small>
							</h1>
						</div>
					</div>
				</div>

				<div class="row">
					<div class="col-md-6">
						<h2>Categories</h2>
						{{if .Categories}}
							<ul id="front-categories">
								{{range $category, $catLink := .Categories}}
									<li><a href="{{ $catLink }}">{{ $category }}</a></li>
								{{end}}
							</ul>
						{{else}}
							<p>No categories are currently being used.</p>
						{{end}}

						<h2>Tags</h2>
						{{if .Tags}}
							<ul id="front-tags">
								{{range $tag, $tagLink := .Tags}}
									<li><a href="{{ $tagLink }}">{{ $tag }}</a></li>
								{{end}}
							</ul>
						{{else}}
							<p>No tags are currently being used.</p>
						{{end}}
					</div>

					<div class="col-md-6">
						<h2>Latest Reports</h2>
						{{if .Reports}}
							<ul id="front-reports">
								{{range .Reports}}
									<li>
                    <a href="{{ .Filename }}">{{ .Title }}</a> &nbsp;
										<span class="details">
											{{if .Category}}
												(<a href="{{ .CategoryURL }}">{{ .Category }}</a>)
											{{end}}
										</span>
                  </li>
								{{end}}
							</ul>

							<p><a href="reports/">See more reports</a></p>
						{{else}}
							<p>There aren't currently any reports.</p>
						{{end}}
					</div>
				</div>

			</div>
		</div>

		<div id="footer" class="container">
			{{ index .Html "footer" }}
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`
