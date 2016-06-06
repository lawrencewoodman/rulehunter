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
	"bytes"
	"fmt"
	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehuntersrv/config"
	"github.com/vlifesystems/rulehuntersrv/html/cmd"
	"github.com/vlifesystems/rulehuntersrv/progress"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// File mode permission used as standard for the html content:
// No special permission bits
// User: Read, Write Execute
// Group: Read
// Other: None
const modePerm = 0740

// This should be run as a goroutine and each time a command is passed to
// cmds the html will be generated
func Run(
	config *config.Config,
	pm *progress.ProgressMonitor,
	logger service.Logger,
	cmds chan cmd.Cmd,
) {
	const minWaitSeconds = 4.0
	lastCmd := cmd.Flush
	lastTime := time.Now()

	go pulse(cmds)
	for c := range cmds {
		durationSinceLast := time.Since(lastTime)
		if c != lastCmd || durationSinceLast.Seconds() > minWaitSeconds {
			if c == cmd.Flush {
				c = lastCmd
				lastCmd = cmd.Flush
			} else {
				lastCmd = c
			}
			lastTime = time.Now()
			if err := generate(c, config, pm); err != nil {
				fullErr := fmt.Errorf("Couldn't generate report: %s", err)
				// TODO: Work out if this is thread safe
				logger.Error(fullErr)
			}
		}
	}
}

// This is used where a command has been received such as when a report has
// finished, but the correct time hasn't elapsed before generating html.
// Therefore a pulse is periodically sent to flush the command.
func pulse(cmds chan cmd.Cmd) {
	sleepInSeconds := time.Duration(4)
	for {
		cmds <- cmd.Flush
		time.Sleep(sleepInSeconds * time.Second)
	}
}

func generate(
	c cmd.Cmd,
	config *config.Config,
	pm *progress.ProgressMonitor,
) error {
	switch c {
	case cmd.Progress:
		if err := generateProgressPage(config, pm); err != nil {
			return err
		}
	case cmd.Reports:
		if err := generateReports(config, pm); err != nil {
			return err
		}
		if err := generateTagPages(config); err != nil {
			return err
		}
		if err := generateProgressPage(config, pm); err != nil {
			return err
		}
	case cmd.All:
		if err := generateHomePage(config, pm); err != nil {
			return err
		}
		if err := generateReports(config, pm); err != nil {
			return err
		}
		if err := generateTagPages(config); err != nil {
			return err
		}
		if err := generateProgressPage(config, pm); err != nil {
			return err
		}
	}
	return nil
}

func inStrings(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

var nonAlphaNumOrSpaceRegexp = regexp.MustCompile("[^[:alnum:] ]")
var spaceRegexp = regexp.MustCompile(" ")
var multipleHyphenRegexp = regexp.MustCompile("-+")

func escapeString(s string) string {
	newS := nonAlphaNumOrSpaceRegexp.ReplaceAllString(s, "")
	newS = spaceRegexp.ReplaceAllString(newS, "-")
	newS = multipleHyphenRegexp.ReplaceAllString(newS, "-")
	newS = strings.Trim(newS, " -")
	return strings.ToLower(newS)
}

func writeTemplate(
	filename string,
	tpl string,
	tplData interface{},
) error {
	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := t.Execute(f, tplData); err != nil {
		return err
	}
	return nil
}

func genReportFilename(wwwDir string, stamp time.Time, title string) string {
	localDir := genReportLocalDir(wwwDir, stamp, title)
	return filepath.Join(localDir, "index.html")
}

// This doesn't change below a second as if two or more reports had the same
// title and were made less than a second apart then you wouldn't be able to
// tell them apart anyway from the list of reports.  This should therfore
// be discouraged.
func genStampMagicString(stamp time.Time) string {
	sum := stamp.Hour()*3600 + stamp.Minute()*60 + stamp.Second()
	return strconv.FormatUint(uint64(sum), 36)
}

func genReportURLDir(
	stamp time.Time,
	title string,
) string {
	magicNumber := genStampMagicString(stamp)
	escapedTitle := escapeString(title)
	return fmt.Sprintf("/reports/%d/%02d/%02d/%s_%s/",
		stamp.Year(), stamp.Month(), stamp.Day(), magicNumber, escapedTitle)
}

func genReportLocalDir(
	wwwDir string,
	stamp time.Time,
	title string,
) string {
	magicNumber := genStampMagicString(stamp)
	escapedTitle := escapeString(title)
	return filepath.Join(wwwDir, "reports",
		fmt.Sprintf("%d", stamp.Year()),
		fmt.Sprintf("%02d", stamp.Month()),
		fmt.Sprintf("%02d", stamp.Day()),
		fmt.Sprintf("%s_%s", magicNumber, escapedTitle))
}

func makeReportURLDir(
	wwwDir string,
	stamp time.Time,
	title string,
) (string, error) {
	URLDir := genReportURLDir(stamp, title)
	localDir := genReportLocalDir(wwwDir, stamp, title)
	err := os.MkdirAll(localDir, modePerm)
	return URLDir, err
}

func countFiles(files []os.FileInfo) int {
	numFiles := 0
	for _, file := range files {
		if !file.IsDir() {
			numFiles++
		}
	}
	return numFiles
}

func makeHtmlNav(menuItem string) template.HTML {
	var doc bytes.Buffer
	validMenuItems := []string{
		"home",
		"reports",
		"tag",
		"progress",
	}

	foundValidItem := false
	for _, validMenuItem := range validMenuItems {
		if validMenuItem == menuItem {
			foundValidItem = true
		}
	}
	if !foundValidItem {
		panic(fmt.Sprintf("menuItem not valid: %s", menuItem))
	}

	t, err := template.New("webpage").Parse(navTpl)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create nav html: %s",
			menuItem, err))
	}

	tplData := struct{ MenuItem string }{menuItem}

	if err := t.Execute(&doc, tplData); err != nil {
		panic(fmt.Sprintf("Couldn't create nav html: %s",
			menuItem, err))
	}
	return template.HTML(doc.String())
}

func makeHtml(menuItem string) map[string]template.HTML {
	r := make(map[string]template.HTML)
	r["head"] = template.HTML(headHtml)
	r["nav"] = makeHtmlNav(menuItem)
	r["bootstrapJS"] = template.HTML(bootstrapJSHtml)
	return r
}
