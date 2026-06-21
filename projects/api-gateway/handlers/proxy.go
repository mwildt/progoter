package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	upstreamURL, err := url.Parse(config.Upstream.URL)
	if err != nil {
		http.Error(w, "Invalid upstream URL", http.StatusInternalServerError)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
	proxy.ServeHTTP(w, r)
}