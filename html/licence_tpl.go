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

const licenceTpl = `
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
				<h1>Licence</h1>
				<p><a href="http://rulehunter.com">Rulehunter</a> - A server to find simple rules in data based on user specified goals.</p>

				<p>Copyright (C) 2016 <a href="http://vlifesystems.com">vLife Systems Ltd</a></p>

				<p>This program is free software: you can redistribute it and/or modify
				it under the terms of the GNU Affero General Public License as published by
				the Free Software Foundation, either version 3 of the License, or
				(at your option) any later version.</p>

				<p>This program is distributed in the hope that it will be useful,
				but WITHOUT ANY WARRANTY; without even the implied warranty of
				MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
				GNU Affero General Public License for more details.</p>

				<p>You should have received a copy of the GNU Affero General Public License
				along with this program.  If not, see
				<a href="http://www.gnu.org/licenses/">http://www.gnu.org/licenses/</a>.</p>

				<h1>Source Code</h1>
				<p>The source code is available at: <a href="{{ .SourceURL }}">{{ .SourceURL }}</a>.</p>

			</div>
		</div>

		<div id="footer" class="container">
			{{ index .Html "footer" }}
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`
