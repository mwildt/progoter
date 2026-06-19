package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticController handles requests for static resources.
type StaticController struct {}

// SetupRoutes configures the routes for static resources.
func (sc *StaticController) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/resources/", sc.ServeStatic)
	mux.HandleFunc("/script.js", sc.ServeScript)
	mux.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/resources/index.html")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/resources/index.html")
	})
}

// ServeScript serves the script.js file with the correct MIME type.
func (sc *StaticController) ServeScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeFile(w, r, "web/resources/script.js")
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

	// Set the Content-Type header based on the file extension
	sc.setContentType(w, fullPath)

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// setContentType sets the Content-Type header based on the file extension.
func (sc *StaticController) setContentType(w http.ResponseWriter, filePath string) {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}