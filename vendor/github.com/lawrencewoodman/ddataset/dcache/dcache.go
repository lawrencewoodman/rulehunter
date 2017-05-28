/*
 * A Go package to cache access to a Dataset
 *
 * Copyright (C) 2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dcache caches a Dataset
package dcache

import (
	"github.com/lawrencewoodman/ddataset"
	"sync"
)

// DCache represents a cached Dataset
type DCache struct {
	dataset      ddataset.Dataset
	cache        []ddataset.Record
	maxCacheRows int
	allCached    bool
	cachedRows   int
	mutex        *sync.Mutex
}

// DCacheConn represents a connection to a DCache Dataset
type DCacheConn struct {
	dataset   *DCache
	conn      ddataset.Conn
	recordNum int
	err       error
}

// New creates a new DCache Dataset which will store up to maxCacheRows
// of another Dataset in memory.  There is only a speed increase if
// maxCacheRows is at least as big as the number of rows in the Dataset
// you want to cache.  However, if it is less the access time is about
// the same.
func New(dataset ddataset.Dataset, maxCacheRows int) ddataset.Dataset {
	return &DCache{
		dataset:      dataset,
		maxCacheRows: maxCacheRows,
		allCached:    false,
		cachedRows:   0,
		mutex:        &sync.Mutex{},
	}
}

// Open creates a connection to the Dataset
func (c *DCache) Open() (ddataset.Conn, error) {
	if c.allCached {
		return &DCacheConn{
			dataset:   c,
			conn:      nil,
			recordNum: -1,
			err:       nil,
		}, nil
	}
	conn, err := c.dataset.Open()
	if err != nil {
		return nil, err
	}
	return &DCacheConn{
		dataset:   c,
		conn:      conn,
		recordNum: -1,
		err:       nil,
	}, nil
}

// Fields returns the field names used by the Dataset
func (c *DCache) Fields() []string {
	return c.dataset.Fields()
}

// Next returns whether there is a Record to be Read
func (cc *DCacheConn) Next() bool {
	if cc.dataset.allCached {
		if cc.recordNum < (cc.dataset.cachedRows - 1) {
			cc.recordNum++
			return true
		}
		return false
	}
	if cc.conn.Err() != nil {
		return false
	}

	isRecord := cc.conn.Next()
	if isRecord {
		cc.recordNum++
	} else if cc.dataset.cachedRows-1 == cc.recordNum && cc.conn.Err() == nil {
		cc.dataset.allCached = true
		cc.conn.Close()
	}

	return isRecord
}

// Err returns any errors from the connection
func (cc *DCacheConn) Err() error {
	if cc.dataset.allCached {
		return nil
	}
	return cc.conn.Err()
}

// Read returns the current Record
func (cc *DCacheConn) Read() ddataset.Record {
	cc.dataset.mutex.Lock()
	defer cc.dataset.mutex.Unlock()
	if cc.recordNum < cc.dataset.cachedRows {
		return cc.dataset.cache[cc.recordNum]
	}

	if cc.dataset.cache == nil {
		cc.dataset.cache = make([]ddataset.Record, cc.dataset.maxCacheRows)
	}

	record := cc.conn.Read()
	if cc.recordNum < cc.dataset.maxCacheRows {
		cc.dataset.cache[cc.recordNum] = record.Clone()
		cc.dataset.cachedRows = cc.recordNum + 1
	}

	return record
}

// Close closes the connection
func (cc *DCacheConn) Close() error {
	if cc.dataset.allCached {
		return nil
	}
	return cc.conn.Close()
}
