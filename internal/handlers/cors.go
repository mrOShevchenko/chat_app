package handlers

import "net/http"

var allowedOrigins = map[string]bool{
	"https://myapp.com":     true,
	"https://dev.myapp.com": true,
	"http://localhost:3000": true,
	"http://localhost:":     true,
	"http://127.0.0.1:3000": true,
}

func CheckOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	_, ok := allowedOrigins[origin]
	return ok
}
