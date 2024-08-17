package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/friskywombat/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

// Feed struct with labeled fields for exporint to json
type Feed struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	UserID    uuid.UUID `json:"user_id"`
}

func toFeed(f database.Feed) Feed {
	return Feed{
		ID:        f.ID,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
		Name:      f.Name,
		URL:       f.Url,
		UserID:    f.UserID,
	}
}

func (cfg *apiConfig) newFeed(ctxt context.Context, name string, url string, userID uuid.UUID) (Feed, error) {
	id := uuid.New()
	now := time.Now().UTC()
	param := database.CreateFeedParams{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
		Url:       url,
		UserID:    userID,
	}
	f, err := cfg.DB.CreateFeed(ctxt, param)
	if err != nil {
		return Feed{}, err
	}
	return toFeed(f), nil
}

func (cfg *apiConfig) newFeedHandleFunc(w http.ResponseWriter, r *http.Request) {
	type feedReq struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	req := feedReq{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, 400, "Error parsing JSON: "+err.Error())
		return
	}
	user, ok := r.Context().Value(userKey).(database.User)
	if !ok {
		respondWithError(w, 500, "Failed to retrieve user data from context")
		return
	}
	feed, err := cfg.newFeed(r.Context(), req.Name, req.URL, user.ID)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 201, feed)
}
