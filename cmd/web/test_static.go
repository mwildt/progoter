package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func testStaticController() {
	staticController := &StaticController{}
	mux := http.NewServeMux()
	staticController.SetupRoutes(mux)

	// Test index.html
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	fmt.Printf("Status Code: %d\n", rec.Code)
	fmt.Printf("Content-Type: %s\n", rec.Header().Get("Content-Type"))
	fmt.Printf("Body: %s\n", rec.Body.String())

	// Test script.js
	req = httptest.NewRequest("GET", "/resources/script.js", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	fmt.Printf("Status Code: %d\n", rec.Code)
	fmt.Printf("Content-Type: %s\n", rec.Header().Get("Content-Type"))
	fmt.Printf("Body: %s\n", rec.Body.String())

	// Test non-existent file
	req = httptest.NewRequest("GET", "/resources/nonexistent.txt", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	fmt.Printf("Status Code: %d\n", rec.Code)
}

func main() {
	testStaticController()
}