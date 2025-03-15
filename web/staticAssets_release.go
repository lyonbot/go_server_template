//go:build release

package web

import (
	"embed"
	"io/fs"
)

//go:embed dist
var webFSbase embed.FS

var WebFS fs.FS

func init() {
	if fs, err := fs.Sub(webFSbase, "dist"); err != nil {
		panic(err)
	} else {
		WebFS = fs
	}
}
