// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package progress

import "fmt"

type Experiment struct {
	Filename string   `json:"filename"`
	Title    string   `json:"title"`
	Tags     []string `json:"tags"`
	Category string   `json:"category"`
	Status   *Status  `json:"status"`
}

func newExperiment(
	filename string,
	title string,
	tags []string,
	category string,
) *Experiment {
	return &Experiment{
		Filename: filename,
		Title:    title,
		Tags:     tags,
		Category: category,
		Status:   NewStatus(),
	}
}

func (e *Experiment) String() string {
	return fmt.Sprintf(
		"{filename: %s, title: %s, tags: %v, category: %s, status: %v}",
		e.Filename, e.Title, e.Tags, e.Category, e.Status,
	)
}
