package main

import (
	"net/http"
	"sync/atomic"

	"github.com/MrInfernoe/Chirpy/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DbURL          string
	Platform       string
	TokenSecret    string
	PolkaKey       string
}

type State struct {
	Config *ApiConfig
	DbQ    *database.Queries
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resw http.ResponseWriter, req *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(resw, req)
	})
}
