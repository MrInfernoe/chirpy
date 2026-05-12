package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MrInfernoe/Chirpy/internal/auth"
	"github.com/MrInfernoe/Chirpy/internal/database"
	"github.com/google/uuid"
)

func errServer(resw http.ResponseWriter, errString string, err error) {
	errMessage := fmt.Sprintf("%v: %v\n", errString, err)
	fmt.Println(errMessage)
	resw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resw.WriteHeader(http.StatusInternalServerError)
	resw.Write([]byte(errMessage))
}

func errClient(resw http.ResponseWriter, errString string) {
	// errMessage := fmt.Sprintf("%v: %v\n", errString, err)
	errMessage := errString
	fmt.Println(errMessage)
	resw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resw.WriteHeader(http.StatusBadRequest)
	resw.Write([]byte(errMessage))
}

func endpointHealth(sm *http.ServeMux, s *State) {
	sm.HandleFunc(http.MethodGet+" /api/healthz", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resw.WriteHeader(http.StatusOK)
		resw.Write([]byte("OK"))
	})
}

func endpointMetrics(sm *http.ServeMux, s *State) {
	sm.HandleFunc(http.MethodGet+" /admin/metrics", func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Set("Content-Type", "text/html")
		resw.WriteHeader(http.StatusOK)
		resBody := fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", s.Config.FileserverHits.Load())
		resw.Write([]byte(resBody))
	})
}

func endpointReset(sm *http.ServeMux, s *State) {
	sm.HandleFunc(http.MethodPost+" /admin/reset", func(resw http.ResponseWriter, req *http.Request) {
		if s.Config.Platform != "dev" {
			resw.WriteHeader(http.StatusForbidden)
			return
		}

		s.Config.FileserverHits.Store(0)

		err := s.DbQ.ResetUsers(req.Context())
		if err != nil {
			// errMessage := fmt.Sprintf("could not reset users: %v\n", err)
			// fmt.Println(errMessage)
			// resw.Header().Set("Content-type", "text/plain")
			// resw.WriteHeader(500)
			// resw.Write([]byte(errMessage))
			errServer(resw, "could not reset users", err)
			return
		}

		resw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resw.WriteHeader(http.StatusOK)
		resBody := "Hits reset to 0. Users database reset."
		resw.Write([]byte(resBody))
	})
}

func endpointValidate_depreciated(sm *http.ServeMux, s *State) {
	sm.HandleFunc(http.MethodPost+" /api/validate_chirp", func(resw http.ResponseWriter, req *http.Request) {
		type validateParameters struct {
			Body string `json:"body"`
		}

		type errChirp struct {
			Error string `json:"error"`
		}

		decoder := json.NewDecoder(req.Body)
		var vPs validateParameters
		err := decoder.Decode(&vPs)
		resw.Header().Set("Content-Type", "application/json")

		if err != nil {
			resw.WriteHeader(http.StatusBadRequest)
			resBody := errChirp{Error: err.Error()}
			data, err := json.Marshal(resBody)
			if err != nil {
				fmt.Printf("error encoding response: %v\n", err)
			}
			resw.Write(data)

		} else if len(vPs.Body) > 140 {
			resw.WriteHeader(http.StatusBadRequest)
			resBody := errChirp{Error: "Chirp is too long"}
			data, err := json.Marshal(resBody)
			if err != nil {
				fmt.Printf("error encoding response: %v\n", err)
			}
			resw.Write(data)

		} else if len(vPs.Body) == 0 {
			resw.WriteHeader(http.StatusBadRequest)
			resBody := errChirp{Error: "Chirp is empty"}
			data, err := json.Marshal(resBody)
			if err != nil {
				fmt.Printf("error encoding response: %v\n", err)
			}
			resw.Write(data)

		} else {
			resw.WriteHeader(http.StatusOK)

			reqBody := vPs.Body

			for _, profanity := range []string{"kerfuffle", "sharbert", "fornax"} {
				for {
					profanityIndex := strings.Index(strings.ToLower(reqBody), profanity)
					if profanityIndex == -1 {
						break
					}
					reqBody = reqBody[:profanityIndex] + "****" + reqBody[profanityIndex+len(profanity):]
				}
			}
			resBody := struct {
				Cleaned_Body string `json:"cleaned_body"`
			}{
				Cleaned_Body: reqBody,
			}
			data, err := json.Marshal(resBody)
			if err != nil {
				fmt.Printf("error encoding response: %v\n", err)
			}
			resw.Write(data)
		}
	})
}

func endpointUsers(sm *http.ServeMux, s *State) {

	sm.HandleFunc(http.MethodPost+" /api/users", func(resw http.ResponseWriter, req *http.Request) {

		reqData := database.CreateUserParams{}
		// reqData := struct {
		// 	Email    string `json:"email"`
		// 	Password string `json:"password"`
		// }{}
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&reqData)
		if err != nil {
			// errMessage := fmt.Sprintf("could not decode: %v\n", err)
			// fmt.Println(errMessage)
			// resw.Header().Set("Conte")
			// resw.WriteHeader(500) // server error
			errServer(resw, "could not decode", err)
			return
		}

		if reqData.Password == "" {
			http.Error(resw, "password is empty", http.StatusBadRequest)
		}

		reqData.Password, err = auth.HashPassword(reqData.Password)
		if err != nil {
			errServer(resw, "could not hash password", err)
		}

		createdUser, err := s.DbQ.CreateUser(req.Context(), reqData)
		if err != nil {
			// conflict if (i know almost impossible) duplicate user_id generated
			// retry will pass through
			// fmt.Printf("error creating user: %v\n", err)
			errServer(resw, "could not create user", err)
			return
		}

		resData, err := json.Marshal(&createdUser)
		if err != nil {
			// fmt.Printf("error encoding user: %v\n", err)
			errServer(resw, "could not encode response", err)
			return
		}

		resw.Header().Set("Content-Type", "application/json")
		resw.WriteHeader(http.StatusCreated)
		resw.Write(resData)
	})
}

func endpointChirps(sm *http.ServeMux, s *State) {
	sm.HandleFunc(http.MethodPost+" /api/chirps", func(resw http.ResponseWriter, req *http.Request) {

		// type errChirp struct {
		// 	Error string `json:"error"`
		// }

		var reqData database.CreateChirpParams
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&reqData)
		if err != nil {
			errServer(resw, "could not decode request", err)
			return
		}

		if len(reqData.Body) > 140 {
			errClient(resw, "Chirp too long")
			return
		}
		if len(reqData.Body) == 0 {
			errClient(resw, "Chirp empty")
			return
		}

		reqBody := reqData.Body
		for _, profanity := range []string{"kerfuffle", "sharbert", "fornax"} {
			for {
				profanityIndex := strings.Index(strings.ToLower(reqBody), profanity)
				if profanityIndex == -1 {
					break
				}
				reqBody = reqBody[:profanityIndex] + "****" + reqBody[profanityIndex+len(profanity):]
			}
		}

		// chirpParams := database.CreateChirpParams{Body: reqBody, UserID: reqData.UserID}
		reqData.Body = reqBody
		createdChirp, err := s.DbQ.CreateChirp(req.Context(), reqData)
		if err != nil {
			errServer(resw, "could not create chirp", err)
			return
		}

		resData, err := json.Marshal(&createdChirp)
		if err != nil {
			errServer(resw, "could not encode response", err)
			return
		}

		resw.Header().Set("Content-Type", "application/json")
		resw.WriteHeader(http.StatusCreated)
		resw.Write(resData)
	})
}

func endpointGetChirps(sm *http.ServeMux, s *State) {
	sm.HandleFunc(http.MethodGet+" /api/chirps", func(resw http.ResponseWriter, req *http.Request) {

		chirps, err := s.DbQ.GetChirps(req.Context())
		if err != nil {
			errServer(resw, "could not get chirps", err)
			return
		}

		resData, err := json.Marshal(&chirps)
		if err != nil {
			errServer(resw, "could not encode chirps", err)
			return
		}

		resw.Header().Set("Content-Type", "application/json")
		resw.WriteHeader(http.StatusOK)
		resw.Write(resData)
	})
}

func endpointGetChirp(sm *http.ServeMux, s *State) {
	sm.HandleFunc(http.MethodGet+" /api/chirps/{chirp_id}", func(resw http.ResponseWriter, req *http.Request) {

		// get chirp id from request pattern
		chirp_id_string := req.PathValue("chirp_id")

		// check if chirp id is not empty
		if chirp_id_string == "" {
			errClient(resw, "no chirp id given")
			return
		}

		// convert to uuid
		chirp_id, err := uuid.Parse(chirp_id_string)
		if err != nil {
			errServer(resw, "could not parse string", err)
			return
		}

		// error here --------------------------------------
		// get chirp from database
		chirp, err := s.DbQ.GetChirp(req.Context(), chirp_id)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				http.Error(resw, "chirp not found", http.StatusNotFound)
				return
			}
			errServer(resw, "could not get chirp", err)
			return
		}

		// encode chirp
		resData, err := json.Marshal(&chirp)
		if err != nil {
			errServer(resw, "could not encode response", err)
		}

		// write response
		resw.Header().Set("Content-type", "application/json")
		resw.WriteHeader(http.StatusOK)
		resw.Write(resData)
	})
}

func endpointLogin(sm *http.ServeMux, s *State) {
	sm.HandleFunc(http.MethodPost+" /api/login", func(resw http.ResponseWriter, req *http.Request) {

		var reqData database.CreateUserParams
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&reqData)
		if err != nil {
			errServer(resw, "could not decode: %v\n", err)
			return
		}

		userWithPassword, err := s.DbQ.GetUserWithPassword(req.Context(), reqData.Email)

		// check if err not nil
		if err != nil {
			// check if found from email
			if err.Error() != "not found" {
				http.Error(resw, "incorrect password or email", http.StatusUnauthorized)
				return
			}
			errServer(resw, "could not get user: err", err)
			return
		}

		// check password hash
		same, err := auth.CheckPasswordHash(reqData.Password, userWithPassword.Password)
		if err != nil {
			errServer(resw, "could not check hash", err)
			return
		}

		if !same {
			http.Error(resw, "incorrect password or email", http.StatusUnauthorized)
			return
		}

		userWithoutPassword := database.CreateUserRow{ID: userWithPassword.ID, CreatedAt: userWithPassword.CreatedAt, UpdatedAt: userWithPassword.UpdatedAt, Email: userWithPassword.Email}
		reswData, err := json.Marshal(&userWithoutPassword)
		if err != nil {
			errServer(resw, "could not encode response", err)
			return
		}

		resw.Header().Set("Content-Type", "application/json")
		resw.WriteHeader(http.StatusOK)
		resw.Write(reswData)
	})
}
