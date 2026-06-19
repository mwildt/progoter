package controller

import (
	"net/http"
	"os"
	"path/filepath"
)

// StaticController handles requests for static resources.
type StaticController struct{}

// SetupRoutes configures the routes for static resources.
func (sc *StaticController) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/resources/", sc.ServeStatic)
	mux.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/resources/index.html")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/resources/index.html")
	})
}

// ServeStatic serves static files from the web/resources directory.
func (sc *StaticController) ServeStatic(w http.ResponseWriter, r *http.Request) {
	// Get the path from the URL
	path := r.URL.Path

	// Construct the full path to the file
	fullPath := filepath.Join("web", "resources", path)

	// Check if the file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}
