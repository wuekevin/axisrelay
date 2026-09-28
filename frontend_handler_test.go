package main

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestSPAHandlerKeepsAppBoundaries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	web := fstest.MapFS{
		"index.html":           &fstest.MapFile{Data: []byte("web-shell")},
		"assets/web-abc123.js": &fstest.MapFile{Data: []byte("web-asset")},
	}
	admin := fstest.MapFS{
		"index.html":             &fstest.MapFile{Data: []byte("admin-shell")},
		"assets/admin-abc123.js": &fstest.MapFile{Data: []byte("admin-asset")},
	}

	router := gin.New()
	router.GET("/", newSPAHandler(fs.FS(web), ""))
	router.GET("/console/*filepath", newSPAHandler(fs.FS(web), ""))
	router.GET("/assets/*filepath", newSPAHandler(fs.FS(web), ""))
	router.GET("/admin", newSPAHandler(fs.FS(admin), "/admin"))
	router.GET("/admin/*filepath", newSPAHandler(fs.FS(admin), "/admin"))

	assertResponse(t, router, "/console/profile", http.StatusOK, "web-shell", "no-store, no-cache, must-revalidate")
	assertResponse(t, router, "/admin/settings", http.StatusOK, "admin-shell", "no-store, no-cache, must-revalidate")
	assertResponse(t, router, "/assets/web-abc123.js", http.StatusOK, "web-asset", "public, max-age=31536000, immutable")
	assertResponse(t, router, "/admin/assets/admin-abc123.js", http.StatusOK, "admin-asset", "public, max-age=31536000, immutable")
	assertResponse(t, router, "/assets/missing-abc123.js", http.StatusNotFound, "", "no-store, no-cache, must-revalidate")
	assertResponse(t, router, "/admin/assets/missing-abc123.js", http.StatusNotFound, "", "no-store, no-cache, must-revalidate")
}

func assertResponse(t *testing.T, handler http.Handler, target string, wantStatus int, wantBody, wantCache string) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != wantStatus {
		t.Fatalf("%s status = %d, want %d", target, recorder.Code, wantStatus)
	}
	if recorder.Body.String() != wantBody {
		t.Fatalf("%s body = %q, want %q", target, recorder.Body.String(), wantBody)
	}
	if got := recorder.Header().Get("Cache-Control"); got != wantCache {
		t.Fatalf("%s Cache-Control = %q, want %q", target, got, wantCache)
	}
}
