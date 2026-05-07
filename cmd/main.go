package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resw http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(resw, req)
	})
}

func main() {
	// matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL.
	const rootDir = "./web/static"
	const port = ":8080"
	const filepathExtension = "/app"

	cfg := &apiConfig{}
	cfg.fileserverHits.Store(0)

	serveMux := http.NewServeMux()
	strippedFileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(rootDir)))
	serveMux.Handle("/app/", cfg.middlewareMetricsInc(strippedFileserverHandler))

	handlerHealth(serveMux, cfg)
	handlerMetrics(serveMux, cfg)
	handlerReset(serveMux, cfg)
	handlerValidate(serveMux, cfg)

	server := http.Server{}
	server.Addr = port
	server.Handler = serveMux

	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		fmt.Println("server closed")
	} else {
		fmt.Println(err)
	}
}
