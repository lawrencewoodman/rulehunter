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

package html

import (
	"sort"
	"time"
)

type TplReport struct {
	Title    string
	Tags     map[string]string
	DateTime string
	Filename string
	Stamp    time.Time
}

func newTplReport(
	title string,
	tags map[string]string,
	filename string,
	stamp time.Time,
) *TplReport {
	return &TplReport{
		title,
		tags,
		stamp.Format(time.RFC822),
		filename,
		stamp,
	}
}

// byDate implements sort.Interface for []*TplReport
type byDate []*TplReport

func (r byDate) Len() int { return len(r) }
func (r byDate) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r byDate) Less(i, j int) bool {
	return r[j].Stamp.Before(r[i].Stamp)
}

func sortTplReportsByDate(tplReports []*TplReport) {
	sort.Sort(byDate(tplReports))
}
