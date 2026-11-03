// Package dashboard embeds the offline React application in the executable.
package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist
var assets embed.FS

func Handler() http.Handler { root, _ := fs.Sub(assets, "dist"); return http.FileServer(http.FS(root)) }

// pages are the dashboard's page addresses (/integrate, /changes...). Each
// serves the dashboard, which opens that page, so a reload or a shared link
// stays on it.
var pages = map[string]bool{"overview": true, "activity": true, "changes": true, "integrate": true, "providers": true, "routing": true,
	"agents": true, "compression": true, "tasks": true, "settings": true, "manual": true}

// IsPage reports whether path is a dashboard page address.
func IsPage(path string) bool { return pages[strings.Trim(path, "/")] }
