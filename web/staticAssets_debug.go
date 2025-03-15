//go:build !release

package web

import (
	"io/fs"
)

// empty
var WebFS fs.FS
