// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

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
