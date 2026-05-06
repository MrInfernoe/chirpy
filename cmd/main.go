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

	serveMux.HandleFunc(http.MethodGet+" /api/healthz", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		resw.WriteHeader(http.StatusOK)
		resw.Write([]byte("OK"))
	})

	serveMux.HandleFunc(http.MethodGet+" /admin/metrics", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Add("Content-Type", "text/html")
		resw.WriteHeader(http.StatusOK)
		body := fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", cfg.fileserverHits.Load())
		resw.Write([]byte(body))
	})

	serveMux.HandleFunc(http.MethodPost+" /admin/reset", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		resw.WriteHeader(http.StatusOK)
		cfg.fileserverHits.Store(0)
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
