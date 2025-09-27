package app

import (
	"net/http"

	chiCors "github.com/go-chi/cors"
)

var allowedHeaders []string = []string{
	"User-Agent",
	"Content-Type",
	"Accept",
	"Accept-Encoding",
	"Accept-Language",
	"Cache-Control",
	"Connection",
	"DNT",
	"Host",
	"Origin",
	"Pragma",
	"Referer",
	"Authorization",
	"X-User-Id",
	"X-User-Email",
}

var allowedMethods []string = []string{
	http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions, http.MethodHead,
}

var cors = chiCors.Handler(chiCors.Options{
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   allowedMethods,
	AllowedHeaders:   allowedHeaders,
	AllowCredentials: true,
})
