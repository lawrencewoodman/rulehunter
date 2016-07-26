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

package main

import (
	"errors"
	"fmt"
)

var errNoConfigDirArg = errors.New("no -configdir argument")
var errInstallAndServeArg = errors.New("can't have -install and -serve argument")

type errConfigLoad struct {
	filename string
	err      error
}

func (e errConfigLoad) Error() string {
	return fmt.Sprintf("couldn't load configuration %s: %s", e.filename, e.err)
}
