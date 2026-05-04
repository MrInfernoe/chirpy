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
	root := http.Dir(rootDir)
	fileserverHandler := http.FileServer(root)
	strippedFileserverHandler := http.StripPrefix(filepathExtension, fileserverHandler)
	fileserverPattern := filepathExtension + "/"
	serveMux.Handle(fileserverPattern, cfg.middlewareMetricsInc(strippedFileserverHandler))

	serveMux.HandleFunc("/healthz", func(resw http.ResponseWriter, req *http.Request) {
		// write header
		resw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		// status code
		resw.WriteHeader(http.StatusOK)
		// body
		resw.Write([]byte("OK"))
	})

	serveMux.HandleFunc("/metrics", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		resw.WriteHeader(http.StatusOK)
		body := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
		resw.Write([]byte(body))
	})

	serveMux.HandleFunc("/reset", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		resw.WriteHeader(http.StatusOK)
		body := "Hits reset to 0"
		resw.Write([]byte(body))
	})

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
