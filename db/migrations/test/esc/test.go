//go:generate esc -o static.go -pkg esc -no-compress ../assets

package esc

import "net/http"

const DirName = "/assets"

// Assets contains project assets.
var Assets http.FileSystem = FS(false)
