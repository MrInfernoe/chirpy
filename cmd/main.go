package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/MrInfernoe/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQ            *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resw http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(resw, req)
	})
}

func main() {
	const rootDir = "./web/static"
	const port = ":8080"
	const fpExt = "/app"

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("error loading env: %v\n", err)
		os.Exit(1)
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("error opening database: %v\n", err)
		os.Exit(1)
	}

	cfg := &apiConfig{}
	cfg.fileserverHits.Store(0)
	cfg.dbQ = database.New(db)

	serveMux := http.NewServeMux()
	strippedFileserverHandler := http.StripPrefix(fpExt, http.FileServer(http.Dir(rootDir)))
	serveMux.Handle(fpExt+"/", cfg.middlewareMetricsInc(strippedFileserverHandler))

	endpointHealth(serveMux, cfg)
	endpointMetrics(serveMux, cfg)
	endpointReset(serveMux, cfg)
	endpointValidate(serveMux, cfg)

	server := http.Server{}
	server.Addr = port
	server.Handler = serveMux

	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		fmt.Println("server closed")
	} else {
		fmt.Printf("error from listening or serving: %v\n", err)
		os.Exit(1)
	}
}
