// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

const tagTpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<title>Reports for tag: {{ .Tag }}</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Reports for tag: {{ .Tag }}</h1>

				<ul class="reports">
					{{range .Reports}}
						<li>
							<a class="title" href="{{ .Filename }}">{{ .Title }}</a><br />
							Date: {{ .DateTime }} &nbsp;
							{{if .Category}}
								Category: <a href="{{ .CategoryURL }}">{{ .Category }}</a> &nbsp;
							{{end}}
							Tags:
							{{range $tag, $catLink := .Tags}}
								<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
							{{end}}
						</li>
					{{end}}
				</ul>
			</div>
		</div>

		<div id="footer" class="container">
			{{ index .Html "footer" }}
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`
