// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package internal

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

func MakeBuildFilename(category, title string) string {
	srcStr := fmt.Sprintf("%s!!%s", category, title)
	hash := sha512.Sum512([]byte(srcStr))
	return string(hex.EncodeToString(hash[:])) + ".json"
}
