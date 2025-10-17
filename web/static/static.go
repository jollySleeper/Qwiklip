package static

import (
	"embed"
	"net/http"
)

//go:embed svg/*.svg css/*.css
var staticFiles embed.FS

// GetStaticFS returns the embedded static file system
func GetStaticFS() http.FileSystem {
	return http.FS(staticFiles)
}
