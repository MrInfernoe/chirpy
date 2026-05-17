package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/MrInfernoe/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	const rootDir = "./web/static"
	const port = ":8080"
	const fpExt = "/app"

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("could not load env: %v\n", err)
		os.Exit(1)
	}

	state := &State{}
	state.Config = &ApiConfig{}
	state.DbQ = &database.Queries{}

	state.Config.Platform = os.Getenv("PLATFORM")
	state.Config.DbURL = os.Getenv("DB_URL")
	state.Config.TokenSecret = os.Getenv("TOKEN_SECRET")
	state.Config.PolkaKey = os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", state.Config.DbURL)
	if err != nil {
		fmt.Printf("could not open database: %v\n", err)
		os.Exit(1)
	}

	state.Config.FileserverHits.Store(0)
	state.DbQ = database.New(db)

	serveMux := http.NewServeMux()
	strippedFileserverHandler := http.StripPrefix(fpExt, http.FileServer(http.Dir(rootDir)))
	serveMux.Handle(fpExt+"/", state.Config.MiddlewareMetricsInc(strippedFileserverHandler))

	endpointHealth(serveMux, state)
	endpointMetrics(serveMux, state)
	endpointReset(serveMux, state)
	// endpointValidate(serveMux, state)
	endpointUsers(serveMux, state)
	endpointChirps(serveMux, state)
	endpointGetChirps(serveMux, state)
	endpointGetChirp(serveMux, state)
	endpointLogin(serveMux, state)
	endpointRefresh(serveMux, state)
	endpointRevoke(serveMux, state)
	endpointUsersUpdate(serveMux, state)
	endpointDeleteChirp(serveMux, state)
	endpointPolkaWebhook(serveMux, state)

	server := http.Server{}
	server.Addr = port
	server.Handler = serveMux

	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		fmt.Println("server closed")
	} else {
		fmt.Printf("could not inititate handling: %v\n", err)
		os.Exit(1)
	}
}
