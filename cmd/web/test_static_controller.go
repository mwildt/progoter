package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticController(t *testing.T) {
	staticController := &StaticController{}
	mux := http.NewServeMux()
	staticController.SetupRoutes(mux)

	// Test index.html
	t.Run("TestIndexHTML", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
		}

		expectedContentType := "text/html; charset=utf-8"
		contentType := rec.Header().Get("Content-Type")
		if contentType != expectedContentType {
			t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
		}
	})

	// Test script.js
	t.Run("TestScriptJS", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/resources/script.js", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
		}

		expectedContentType := "application/javascript; charset=utf-8"
		contentType := rec.Header().Get("Content-Type")
		if contentType != expectedContentType {
			t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
		}
	})

	// Test non-existent file
	t.Run("TestNonExistentFile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/resources/nonexistent.txt", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}