package httpPlus

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
)

// GetResourcePath returns the resoure path for the server.
func GetResourcePath(baseDir string) (path string) {
	path, err := filepath.Abs(filepath.Join(baseDir, "resources"))
	if err != nil {
		panic(err)
	}

	return
}

// SetupTestServer sets up the test server.
func SetupTestServer(baseDir string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(GetHTTPHandler(baseDir)))
}

// GetHTTPHandler returns the HTTP handle used for server the files.
func GetHTTPHandler(baseDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		args := append([]string{GetResourcePath(baseDir)},
			strings.Split(r.URL.String(), "/")...)
		path := filepath.Join(args...)
		if file, err := os.Open(path); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			io.Copy(w, file)
		}
	}
}
