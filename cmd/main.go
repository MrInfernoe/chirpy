package main

import (
	"fmt"
	"net/http"
)

// func readinessFunc(resW http.ResponseWriter, req *http.Request) {

// }

func main() {
	// matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL.
	const rootDir = "./web/static"
	const port = ":8080"
	const filepathExtension = "/app"

	// cfg := apiconfig.NewApiConfig()

	serveMux := http.NewServeMux()
	root := http.Dir(rootDir)
	fileserverHandler := http.FileServer(root)
	strippedFileserverHandler := http.StripPrefix(filepathExtension, fileserverHandler)
	fileserverPattern := filepathExtension + "/"
	serveMux.Handle(fileserverPattern /*cfg.MiddlewareMetricsInc(*/, strippedFileserverHandler) //)

	serveMux.HandleFunc("/healthz", func(resw http.ResponseWriter, req *http.Request) {
		// write header
		resw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		// status code
		resw.WriteHeader(http.StatusOK)
		// body
		resw.Write([]byte("OK"))
	})

	// serveMux.HandleFunc("/metrics", cfg.MetricsHandler)

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
