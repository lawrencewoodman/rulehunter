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

				{{if eq .MenuItem "licence"}}
					<li class="active"><a href="licence/">Licence</a></li>
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
