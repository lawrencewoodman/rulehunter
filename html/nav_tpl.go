// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

const navTpl = `
<nav class="navbar navbar-inverse navbar-fixed-top">
	<div class="container">
		<div class="navbar-header">
			<button type="button" class="navbar-toggle collapsed"
			        data-toggle="collapse" data-target="#navbar"
			        aria-expanded="false" aria-controls="navbar">
				<span class="sr-only">Toggle navigation</span>
				<span class="icon-bar"></span>
				<span class="icon-bar"></span>
				<span class="icon-bar"></span>
			</button>
			{{ if eq .MenuItem "front"}}
				<a class="navbar-brand active" href="">Rulehunter</a>
			{{else}}
				<a class="navbar-brand" href="">Rulehunter</a>
			{{end}}
		</div>

		<div id="navbar" class="collapse navbar-collapse">
			<ul class="nav navbar-nav">
				{{if eq .MenuItem "reports"}}
					<li class="active"><a href="reports/">Reports</a></li>
				{{else}}
					<li><a href="reports/">Reports</a></li>
				{{end}}

				{{if eq .MenuItem "category"}}
					<li class="active"><a href=".">Category</a></li>
				{{end}}

				{{if eq .MenuItem "tag"}}
					<li class="active"><a href=".">Tag</a></li>
				{{end}}

				{{if eq .MenuItem "activity"}}
					<li class="active"><a href="activity/">Activity</a></li>
				{{else}}
					<li><a href="activity/">Activity</a></li>
				{{end}}
			</ul>
		</div><!--/.nav-collapse -->
	</div>
</nav>`
