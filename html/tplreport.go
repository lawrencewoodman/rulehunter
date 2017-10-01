// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

import (
	"sort"
	"time"
)

type TplReport struct {
	Title       string
	Tags        map[string]string
	Category    string
	CategoryURL string
	DateTime    string
	Filename    string
	Stamp       time.Time
}

func newTplReport(
	title string,
	tags map[string]string,
	category string,
	categoryURL string,
	filename string,
	stamp time.Time,
) *TplReport {
	return &TplReport{
		Title:       title,
		Tags:        tags,
		Category:    category,
		CategoryURL: categoryURL,
		DateTime:    stamp.Format(time.RFC822),
		Filename:    filename,
		Stamp:       stamp,
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
