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

package progress

import "fmt"

type Experiment struct {
	Filename string   `json:"filename"`
	Title    string   `json:"title"`
	Tags     []string `json:"tags"`
	Status   *Status  `json:"status"`
}

func newExperiment(filename string, title string, tags []string) *Experiment {
	return &Experiment{
		Filename: filename,
		Title:    title,
		Tags:     tags,
		Status:   NewStatus(),
	}
}

func (e *Experiment) String() string {
	return fmt.Sprintf("{Filename: %s, Title: %s, Tags: %v, Status: %v}",
		e.Filename, e.Title, e.Tags, e.Status)
}
