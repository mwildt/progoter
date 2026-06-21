package main

import (
	"net/http"
	"api-gateway/handlers"
)

func main() {
	http.Handle("/login", http.HandlerFunc(handlers.LoginHandler))
	http.Handle("/", handlers.AuthMiddleware(http.HandlerFunc(handlers.ProxyHandler)))
	http.ListenAndServe(":8080", nil)
}