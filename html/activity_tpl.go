/*
	rulehunter - A server to find rules in data based on user specified goals
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

const activityTpl = `
<!DOCTYPE html>
<html>
  <head>
		{{ index .Html "head" }}
    <title>Activity</title>
  </head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Activity</h1>

				<div id="reports-container">
					<ul class="reports-activity">
					{{range .Experiments}}
						<li>
							<table class="table table-bordered">
								<tr>
									<th>Date</th>
									<td>{{ .Stamp }}</td>
								</tr>
								{{if .Title}}
									<tr><th>Title</th><td>{{ .Title }}</td></tr>
								{{end}}
								{{if .Category}}
									<tr>
										<th>Category</th>
										<td><a href="{{ .CategoryURL }}">{{ .Category }}</a></td>
									</tr>
								{{end}}
								{{if .Tags}}
									<tr>
										<th>Tags</th>
										<td>
											{{if eq .Status "success"}}
												{{range $tag, $catLink := .Tags}}
													<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
												{{end}}
											{{else}}
												{{range $tag, $catLink := .Tags}}
													{{ $tag }} &nbsp;
												{{end}}
											{{end}}
										</td>
									</tr>
								{{end}}
								<tr><th>Experiment filename</th><td>{{ .Filename }}</td></tr>
								<tr>
									<th>Message</th>
									{{if gt .Percent 0.0}}
										<td>
											{{ .Msg }}<br />
											<progress max="100" value="{{ .Percent }}">
												Progress: {{ .Percent }}%
											</progress>
										</td>
									{{else}}
										<td>{{ .Msg }}</td>
									{{end}}
								</tr>
								<tr>
									<th>Status</th>
									<td class="status-{{ .Status }}">{{ .Status | ToTitle }}</td>
								</tr>
							</table>
						</li>
					{{end}}
					</ul>
				</div>
			</div>
		</div>

		<div id="footer" class="container">
			{{ index .Html "footer" }}
		</div>

		{{ index .Html "bootstrapJS" }}

		<script>
			(function refreshWorker(){
					// Don't cache ajax or content won't fresh
					$.ajaxSetup ({
							cache: false,
							complete: function() {
								// Schedule next request when current one is complete
								setTimeout(refreshWorker, 10000);
							}
					});
					var ajaxLoad = "<img src='img/ring.gif' style='width:48px; height:48px' alt='loading...' />";
					var loadUrl = "activity #reports-container";
					$("#reports-container").html(ajaxLoad).load(loadUrl);
			})();
		</script>
	</body>
</html>`
