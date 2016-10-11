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

const homeTpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<meta charset="UTF-8">
		<title>Rulehunter</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Rulehunter</h1>
				Find simple rules in your data to meet your goals.

				<h2>Source</h2>
				Copyright (C) 2016 <a href="http://vlifesystems.com">vLife Systems Ltd</a>

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

				<p>The source code is available at: <a href="{{ .SourceURL }}">{{ .SourceURL }}</a>.</p>

			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`
