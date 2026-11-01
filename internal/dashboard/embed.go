// Package dashboard embeds the offline React application in the executable.
package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var assets embed.FS

func Handler() http.Handler { root, _ := fs.Sub(assets, "dist"); return http.FileServer(http.FS(root)) }
