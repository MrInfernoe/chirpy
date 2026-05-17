package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

		type reqFields struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		var reqData reqFields
		// reqData := database.CreateUserParams{}
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
			return
		}

		reqData.Password, err = auth.HashPassword(reqData.Password)
		if err != nil {
			errServer(resw, "could not hash password", err)
			return
		}

		reqUser := database.CreateUserParams{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Email:     reqData.Email,
			Password:  reqData.Password,
		}

		userWithoutPassword, err := s.DbQ.CreateUser(req.Context(), reqUser)
		if err != nil {
			// conflict if (i know almost impossible) duplicate user_id generated
			// retry will pass through
			// fmt.Printf("error creating user: %v\n", err)
			errServer(resw, "could not create user", err)
			return
		}

		resData, err := json.Marshal(&userWithoutPassword)
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

		bearerToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			errClient(resw, err.Error())
			return
		}

		type reqFields struct {
			Body string `json:"body"`
		}

		var reqData reqFields
		decoder := json.NewDecoder(req.Body)
		err = decoder.Decode(&reqData)
		if err != nil {
			errServer(resw, "could not decode request", err)
			return
		}

		userId, err := auth.ValidateJWT(bearerToken, s.Config.TokenSecret)
		if err != nil {
			http.Error(resw, "401 Unauthorized", http.StatusUnauthorized)
			// errServer(resw, "could not validate", err)
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

		chirpParams := database.CreateChirpParams{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Body:      reqBody,
			UserID:    userId,
		}

		createdChirp, err := s.DbQ.CreateChirp(req.Context(), chirpParams)
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

		type reqFields struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		var reqData reqFields
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
			if err.Error() == "not found" {
				http.Error(resw, "incorrect password or email", http.StatusUnauthorized)
				return
			}
			errServer(resw, "could not get user", err)
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

		tokenString, err := auth.MakeJWT(userWithPassword.ID, s.Config.TokenSecret)
		if err != nil {
			errServer(resw, "could not create token", err)
			return
		}

		refreshToken := auth.MakeRefreshToken()

		RTParams := database.CreateRefreshTokenParams{
			Token:     refreshToken,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			UserID:    userWithPassword.ID,
			ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
		}

		createdRefreshToken, err := s.DbQ.CreateRefreshToken(req.Context(), RTParams)
		if err != nil {
			errServer(resw, "could not add refresh token to database", err)
			return
		}

		type resFields struct {
			ID           uuid.UUID `json:"id"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			Email        string    `json:"email"`
			Token        string    `json:"token"`
			RefreshToken string    `json:"refresh_token"`
			IsChirpyRed  bool      `json:"is_chirpy_red"`
		}

		userWithoutPassword := resFields{
			ID:           userWithPassword.ID,
			CreatedAt:    userWithPassword.CreatedAt,
			UpdatedAt:    userWithPassword.UpdatedAt,
			Email:        userWithPassword.Email,
			Token:        tokenString,
			RefreshToken: createdRefreshToken.Token,
			IsChirpyRed:  userWithPassword.IsChirpyRed,
		}

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

func endpointRefresh(sm *http.ServeMux, s *State) {

	sm.HandleFunc(http.MethodPost+" /api/refresh", func(resw http.ResponseWriter, req *http.Request) {

		bearerToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			errClient(resw, err.Error())
			return
		}

		foundRefreshToken, err := s.DbQ.GetRefreshToken(req.Context(), bearerToken)
		if err != nil {
			http.Error(resw, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		currentTime := time.Now().UTC()
		expiredTime := foundRefreshToken.ExpiresAt.UTC()
		revokedTime := foundRefreshToken.RevokedAt.Time.UTC()
		isRevoked := foundRefreshToken.RevokedAt.Valid
		if currentTime.After(expiredTime) || (isRevoked && currentTime.After(revokedTime)) {
			http.Error(resw, "401 Unauthorized", http.StatusUnauthorized)
			return
		}
		// if isRevoked && currentTime.After(revokedTime) {
		// 	http.Error(resw, "401 Unauthorized", http.StatusUnauthorized)
		// 	return
		// }

		createdAccessToken, err := auth.MakeJWT(foundRefreshToken.UserID, s.Config.TokenSecret)
		if err != nil {
			errServer(resw, "could not create access token", err)
		}

		type reswFields struct {
			AccessToken string `json:"token"`
		}

		accessToken := reswFields{
			AccessToken: createdAccessToken,
		}

		reswData, err := json.Marshal(&accessToken)
		if err != nil {
			errServer(resw, "could not encode response", err)
			return
		}

		resw.Header().Set("Content-Type", "application/json")
		resw.WriteHeader(http.StatusOK)
		resw.Write(reswData)
	})
}

func endpointRevoke(sm *http.ServeMux, s *State) {

	sm.HandleFunc(http.MethodPost+" /api/revoke", func(resw http.ResponseWriter, req *http.Request) {

		bearerToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			errClient(resw, err.Error())
			return
		}

		revokeParams := database.RevokeTokenParams{
			Token: bearerToken,
			RevokedAt: sql.NullTime{
				Time:  time.Now().UTC(),
				Valid: true,
			},
		}

		_, err = s.DbQ.RevokeToken(req.Context(), revokeParams)
		if err != nil {
			errServer(resw, "could not revoke token", err)
			return
		}

		resw.WriteHeader(http.StatusNoContent)
	})
}

func endpointUsersUpdate(sm *http.ServeMux, s *State) {

	sm.HandleFunc(http.MethodPut+" /api/users", func(resw http.ResponseWriter, req *http.Request) {

		bearerToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			if err.Error() == "no authorization token found" {
				http.Error(resw, "401 Unauthorized", http.StatusUnauthorized)
				return
			}
			errClient(resw, err.Error())
			return
		}

		userId, err := auth.ValidateJWT(bearerToken, s.Config.TokenSecret)
		if err != nil {
			http.Error(resw, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		type reqFields struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		var reqData reqFields
		decoder := json.NewDecoder(req.Body)
		err = decoder.Decode(&reqData)
		if err != nil {
			errServer(resw, "could not decode: %v\n", err)
			return
		}

		if reqData.Email == "" {
			http.Error(resw, "email is empty", http.StatusBadRequest)
			return
		}
		if reqData.Password == "" {
			http.Error(resw, "password is empty", http.StatusBadRequest)
			return
		}
		hashword, err := auth.HashPassword(reqData.Password)
		if err != nil {
			errServer(resw, "could not hash password", err)
			return
		}

		userParams := database.UpdateEmailPasswordParams{
			ID:        userId,
			Email:     reqData.Email,
			Password:  hashword,
			UpdatedAt: time.Now().UTC(),
		}

		updatedUserWithoutPassword, err := s.DbQ.UpdateEmailPassword(req.Context(), userParams)
		if err != nil {
			errServer(resw, "could not update database", err)
			return
		}

		reswBytes, err := json.Marshal(&updatedUserWithoutPassword)
		if err != nil {
			errServer(resw, "could not encode response", err)
			return
		}

		resw.Header().Set("Content-Type", "application/json")
		resw.WriteHeader(http.StatusOK)
		resw.Write(reswBytes)
	})
}

func endpointDeleteChirp(sm *http.ServeMux, s *State) {
	sm.HandleFunc(http.MethodDelete+" /api/chirps/{chirp_id}", func(resw http.ResponseWriter, req *http.Request) {

		bearerToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(resw, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenUserId, err := auth.ValidateJWT(bearerToken, s.Config.TokenSecret)
		if err != nil {
			errClient(resw, "could not validate token")
			return
		}

		// get chirp id from request pattern
		chirp_id_string := req.PathValue("chirp_id")

		// check if chirp id is not empty
		if chirp_id_string == "" {
			errClient(resw, "no chirp id given")
			return
		}

		// convert to chirp uuid
		chirpId, err := uuid.Parse(chirp_id_string)
		if err != nil {
			errServer(resw, "could not parse string", err)
			return
		}

		// get chirp from database
		chirp, err := s.DbQ.GetChirp(req.Context(), chirpId)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				http.Error(resw, "chirp not found", http.StatusNotFound)
				return
			}
			errServer(resw, "could not get chirp", err)
			return
		}

		if chirp.UserID != tokenUserId {
			http.Error(resw, "403 Forbidden", http.StatusForbidden)
			return
		}

		delChirpParams := database.DeleteChirpParams{
			ID:     chirp.ID,
			UserID: tokenUserId,
		}
		_, err = s.DbQ.DeleteChirp(req.Context(), delChirpParams)
		if err != nil {
			errServer(resw, "could not delete chirp", err)
			return
		}

		// encode chirp
		// resData, err := json.Marshal(&deletedChirp)
		// if err != nil {
		// 	errServer(resw, "could not encode response", err)
		// }

		// write response
		resw.WriteHeader(http.StatusNoContent)
		// resw.Write(resData)
	})
}

func endpointPolkaWebhook(sm *http.ServeMux, s *State) {

	sm.HandleFunc(http.MethodPost+" /api/polka/webhooks", func(resw http.ResponseWriter, req *http.Request) {

		type reqFields struct {
			Event string `json:"event"`
			Data  struct {
				UserId string `json:"user_id"`
			} `json:"data"`
		}

		var reqData reqFields
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&reqData)
		if err != nil {
			errServer(resw, "could not decode request", err)
			return
		}

		if reqData.Event != "user.upgraded" {
			http.Error(resw, "", http.StatusNoContent)
			return
		} else {
			userId, err := uuid.Parse(reqData.Data.UserId)
			if err != nil {
				errServer(resw, "could not parse user id", err)
				return
			}
			redParams := database.UpgradeToRedParams{
				ID:        userId,
				UpdatedAt: time.Now().UTC(),
			}
			err = s.DbQ.UpgradeToRed(req.Context(), redParams)
			if err != nil {
				if err.Error() == "sql: no rows in result set" {
					http.Error(resw, "user not found", http.StatusNotFound)
					return
				}
				errServer(resw, "could not upgrade user to red in database", err)
				return
			}
		}

		resw.WriteHeader(http.StatusNoContent)
	})
}
