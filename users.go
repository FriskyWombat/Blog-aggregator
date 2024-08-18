package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/friskywombat/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

// User type with labeled fields for marshaling json
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
}

func toUser(u database.User) User {
	return User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Name:      u.Name,
		APIKey:    u.ApiKey,
	}
}

func (cfg *apiConfig) newUser(ctxt context.Context, name string) (User, error) {
	id := uuid.New()
	param := database.CreateUserParams{
		ID:   id,
		Name: name,
	}
	u, err := cfg.DB.CreateUser(ctxt, param)
	if err != nil {
		return User{}, err
	}
	return toUser(u), nil
}

func (cfg *apiConfig) newUserHandleFunc(w http.ResponseWriter, r *http.Request) {
	type userReq struct {
		Name string `json:"name"`
	}
	req := userReq{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, 400, "Error parsing JSON: "+err.Error())
		return
	}
	user, err := cfg.newUser(r.Context(), req.Name)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 201, user)
}

func (cfg *apiConfig) getUserHandleFunc(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey).(database.User)
	if !ok {
		log.Println("getUserHandleFunc has junk instead of user")
		respondWithError(w, 500, "Failed to retrieve user data from context")
		return
	}
	respondWithJSON(w, 200, toUser(user))
}
