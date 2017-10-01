// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

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

						<p><a href="reports/nocategory/">See uncategorized reports</a></p>
						<br />

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

				<div class="row">
					<div class="col-md-12">
						<h2>Activity</h2>
						<div id="front-activity-container">
							{{if .Experiments}}
								{{range .Experiments}}
									{{if .Title}}
										<strong>{{ .Title }}</strong>
									{{end}}
									{{if .Category}}
										&nbsp ({{ .Category }})
									{{end}}
									<br />
									<span class="status-{{ .Status }}">{{ .Status | ToTitle }}</span> &nbsp;
									{{if gt .Percent 0.0}}
										{{ .Msg }}<br />
										<progress max="100" value="{{ .Percent }}">
											Progress: {{ .Percent }}%
										</progress>
									{{else}}
										{{ .Msg }}
									{{end}}
								{{end}}
							{{else}}
							<p>Not currently processing any experiments.</p>
							{{end}}
						</div>
					</div>
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
					var loadUrl = " #front-activity-container";
					$("#front-activity-container").html(ajaxLoad).load(loadUrl);
			})();
		</script>
	</body>
</html>`
