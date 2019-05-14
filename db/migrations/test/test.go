//go:generate esc -o static.go -pkg test -no-compress assets

package test

import "net/http"

const DirName = "/assets"

// Assets contains project assets.
var Assets http.FileSystem = FS(false)
