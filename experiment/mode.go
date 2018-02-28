// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package experiment

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcopy"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/lawrencewoodman/ddataset/dsql"
	"github.com/lawrencewoodman/ddataset/dtruncate"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/fileinfo"
)

type Mode struct {
	Dataset ddataset.Dataset
	When    *dexpr.Expr
}

type InvalidWhenExprError string

func (e InvalidWhenExprError) Error() string {
	return "when field invalid: " + string(e)
}

type modeDesc struct {
	Dataset *datasetDesc `yaml:"dataset"`
	// An expression that works out whether to run the experiment for this mode
	When string `yaml:"when"`
}

type datasetDesc struct {
	CSV    *csvDesc `yaml:"csv"`
	SQL    *sqlDesc `yaml:"sql"`
	Fields []string `yaml:"fields"`
}

type csvDesc struct {
	Filename  string `yaml:"filename"`
	HasHeader bool   `yaml:"hasHeader"`
	Separator string `yaml:"separator"`
}

type sqlDesc struct {
	DriverName     string `yaml:"driverName"`
	DataSourceName string `yaml:"dataSourceName"`
	Query          string `yaml:"query"`
}

func newMode(
	modeField string,
	cfg *config.Config,
	desc *modeDesc,
) (*Mode, error) {
	d, err := makeDataset(cfg, desc.Dataset)
	if err != nil {
		return nil, fmt.Errorf("dataset: %s", err)
	}
	when, err := makeWhenExpr(desc.When)
	if err != nil {
		return nil, InvalidWhenExprError(desc.When)
	}
	return &Mode{
		Dataset: d,
		When:    when,
	}, nil
}

func (m *Mode) ShouldProcess(
	file fileinfo.FileInfo,
	isFinished bool,
	pmStamp time.Time,
) (bool, error) {
	if isFinished && file.ModTime().After(pmStamp) {
		isFinished, pmStamp = false, time.Now()
	}
	return evalWhenExpr(time.Now(), isFinished, pmStamp, m.When)
}

func makeDataset(
	cfg *config.Config,
	dd *datasetDesc,
) (ddataset.Dataset, error) {
	// File mode permission:
	// No special permission bits
	// User: Read, Write Execute
	// Group: None
	// Other: None
	const modePerm = 0700
	var dataset ddataset.Dataset
	if dd.CSV != nil && dd.SQL != nil {
		return nil, errors.New("can't specify csv and sql source")
	}
	if dd.CSV == nil && dd.SQL == nil {
		return nil,
			errors.New("has no csv or sql field")
	}
	if dd.CSV != nil {
		if dd.CSV.Filename == "" {
			return nil, errors.New("csv: missing filename")
		}
		if dd.CSV.Separator == "" {
			return nil, errors.New("csv: missing separator")
		}
		dataset = dcsv.New(
			dd.CSV.Filename,
			dd.CSV.HasHeader,
			rune(dd.CSV.Separator[0]),
			dd.Fields,
		)
	} else if dd.SQL != nil {
		if dd.SQL.DriverName == "" {
			return nil, errors.New("sql: missing driverName")
		}
		if dd.SQL.DataSourceName == "" {
			return nil, errors.New("sql: missing dataSourceName")
		}
		if dd.SQL.Query == "" {
			return nil, errors.New("sql: missing query")
		}
		sqlHandler, err := newSQLHandler(
			dd.SQL.DriverName,
			dd.SQL.DataSourceName,
			dd.SQL.Query,
		)
		if err != nil {
			return nil, fmt.Errorf("sql: %s", err)
		}
		dataset = dsql.New(sqlHandler, dd.Fields)
	}

	if cfg.MaxNumRecords >= 1 {
		dataset = dtruncate.New(dataset, cfg.MaxNumRecords)
	}
	// Copy dataset to get stable version
	buildTmpDir := filepath.Join(cfg.BuildDir, "tmp")
	if err := os.MkdirAll(buildTmpDir, modePerm); err != nil {
		return nil, err
	}
	copyDataset, err := dcopy.New(dataset, buildTmpDir)
	if err != nil {
		return nil, err
	}
	return copyDataset, nil
}
