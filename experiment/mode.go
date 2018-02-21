// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package experiment

import (
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

type modeDesc struct {
	Dataset *datasetDesc `yaml:"dataset"`
	// An expression that works out whether to run the experiment for this mode
	When string `yaml:"when"`
}

func newMode(
	modeField string,
	cfg *config.Config,
	fields []string,
	desc *modeDesc,
) (*Mode, error) {
	d, err := makeDataset(modeField+" > dataset", cfg, fields, desc.Dataset)
	if err != nil {
		return nil, err
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
	modeField string,
	cfg *config.Config,
	fields []string,
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
		return nil, fmt.Errorf(
			"Experiment field: %s, can't specify csv and sql source",
			modeField,
		)
	}
	if dd.CSV == nil && dd.SQL == nil {
		return nil,
			fmt.Errorf("Experiment field: %s, has no csv or sql field", modeField)
	}
	if dd.CSV != nil {
		if dd.CSV.Filename == "" {
			return nil, fmt.Errorf("Experiment field missing: %s > csv > filename",
				modeField)
		}
		if dd.CSV.Separator == "" {
			return nil, fmt.Errorf("Experiment field missing: %s > csv > separator",
				modeField)
		}
		dataset = dcsv.New(
			dd.CSV.Filename,
			dd.CSV.HasHeader,
			rune(dd.CSV.Separator[0]),
			fields,
		)
	} else if dd.SQL != nil {
		if dd.SQL.DriverName == "" {
			return nil, fmt.Errorf(
				"Experiment field missing: %s > sql > driverName",
				modeField,
			)
		}
		if dd.SQL.DataSourceName == "" {
			return nil, fmt.Errorf(
				"Experiment field missing: %s > sql > dataSourceName",
				modeField,
			)
		}
		if dd.SQL.Query == "" {
			return nil, fmt.Errorf("Experiment field missing: %s > sql > query",
				modeField)
		}
		sqlHandler, err := newSQLHandler(
			dd.SQL.DriverName,
			dd.SQL.DataSourceName,
			dd.SQL.Query,
		)
		if err != nil {
			return nil, fmt.Errorf("Experiment field: %s > sql, has %s",
				modeField, err)
		}
		dataset = dsql.New(sqlHandler, fields)
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
