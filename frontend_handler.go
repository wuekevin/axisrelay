package main

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// newSPAHandler serves one independently built frontend. routePrefix is
// stripped before looking up assets, which prevents Web and Admin from ever
// falling back to each other's HTML shell.
func newSPAHandler(appFS fs.FS, routePrefix string) gin.HandlerFunc {
	httpFS := http.FS(appFS)
	indexHTML, indexErr := fs.ReadFile(appFS, "index.html")
	prefix := strings.TrimSuffix(routePrefix, "/")

	return func(c *gin.Context) {
		if indexErr != nil {
			c.Status(http.StatusNotFound)
			return
		}

		requestPath := c.Request.URL.Path
		if prefix != "" {
			if requestPath != prefix && !strings.HasPrefix(requestPath, prefix+"/") {
				c.Status(http.StatusNotFound)
				return
			}
			requestPath = strings.TrimPrefix(requestPath, prefix)
		}
		assetPath := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
		if assetPath != "" && assetPath != "." {
			if file, err := appFS.Open(assetPath); err == nil {
				info, statErr := file.Stat()
				_ = file.Close()
				if statErr == nil && !info.IsDir() {
					if strings.HasPrefix(assetPath, "assets/") {
						c.Header("Cache-Control", "public, max-age=31536000, immutable")
					} else {
						c.Header("Cache-Control", "no-cache")
					}
					c.FileFromFS("/"+assetPath, httpFS)
					return
				}
			}
			if strings.HasPrefix(assetPath, "assets/") {
				c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
				c.Status(http.StatusNotFound)
				return
			}
		}

		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		if c.Query("refresh-assets") == "1" {
			c.Header("Clear-Site-Data", `"cache"`)
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}
}
