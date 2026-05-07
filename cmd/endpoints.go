package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func handlerHealth(sm *http.ServeMux, _ *apiConfig) {
	sm.HandleFunc(http.MethodGet+" /api/healthz", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		resw.WriteHeader(http.StatusOK)
		resw.Write([]byte("OK"))
	})
}

func handlerMetrics(sm *http.ServeMux, cfg *apiConfig) {
	sm.HandleFunc(http.MethodGet+" /admin/metrics", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Add("Content-Type", "text/html")
		resw.WriteHeader(http.StatusOK)
		body := fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", cfg.fileserverHits.Load())
		resw.Write([]byte(body))
	})
}

func handlerReset(sm *http.ServeMux, cfg *apiConfig) {
	sm.HandleFunc(http.MethodPost+" /admin/reset", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		resw.WriteHeader(http.StatusOK)
		cfg.fileserverHits.Store(0)
		body := "Hits reset to 0"
		resw.Write([]byte(body))
	})
}

func handlerValidate(sm *http.ServeMux, _ *apiConfig) {
	sm.HandleFunc(http.MethodPost+" /api/validate_chirp", func(resw http.ResponseWriter, req *http.Request) {
		type validateParameters struct {
			Body string `json:"body"`
		}

		type validChirp struct {
			Valid bool `json:"valid"`
		}

		type errChirp struct {
			Error string `json:"error"`
		}

		decoder := json.NewDecoder(req.Body)
		var vPs validateParameters
		err := decoder.Decode(&vPs)
		resw.Header().Add("Content-Type", "application/json")

		if err != nil {
			resw.WriteHeader(http.StatusBadRequest)
			resBody := errChirp{Error: err.Error()}
			data, err := json.Marshal(resBody)
			if err != nil {
				fmt.Printf("error encoding response: %v\n", err)
			}
			resw.Write(data)

			// resw.Write(fmt.Append([]byte{}, fmt.Sprintf("{\n%v\n}", err)))

		} else if len(vPs.Body) > 140 {
			resw.WriteHeader(http.StatusBadRequest)
			resBody := errChirp{Error: "Chirp is too long"}
			data, err := json.Marshal(resBody)
			if err != nil {
				fmt.Printf("error encoding response: %v\n", err)
			}
			resw.Write(data)

			// resw.Write(fmt.Append([]byte{}, "{\n\"error\": \"Chirp is too long\"\n}"))

		} else if len(vPs.Body) == 0 {
			resw.WriteHeader(http.StatusBadRequest)
			resBody := errChirp{Error: "Chirp is empty"}
			data, err := json.Marshal(resBody)
			if err != nil {
				fmt.Printf("error encoding response: %v\n", err)
			}
			resw.Write(data)
			// resw.Write(fmt.Append([]byte{}, "{\n\"error\": \"Chirp is empty\"\n}"))

		} else {
			resw.WriteHeader(http.StatusOK)
			resBody := validChirp{Valid: true}
			data, err := json.Marshal(resBody)
			if err != nil {
				fmt.Printf("error encoding response: %v\n", err)
			}
			resw.Write(data)
		}
	})
}
