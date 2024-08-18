package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/friskywombat/blog-aggregator/internal/database"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Follow struct {
	ID        uuid.UUID `json:"id"`
	FeedID    uuid.UUID `json:"feed_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toFollow(f database.Follow) Follow {
	return Follow{
		ID:        f.ID,
		FeedID:    f.FeedID,
		UserID:    f.UserID,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func (cfg *apiConfig) newFollow(ctxt context.Context, feedID uuid.UUID, userID uuid.UUID) (Follow, error) {
	id := uuid.New()
	now := time.Now().UTC()
	param := database.CreateFollowParams{
		ID:        id,
		UserID:    userID,
		FeedID:    feedID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	f, err := cfg.DB.CreateFollow(ctxt, param)
	if err != nil {
		return Follow{}, err
	}
	return toFollow(f), nil
}

func (cfg *apiConfig) newFollowHandleFunc(w http.ResponseWriter, r *http.Request) {
	type feedReq struct {
		FeedID uuid.UUID `json:"feed_id"`
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
	feed, err := cfg.newFollow(r.Context(), req.FeedID, user.ID)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 201, feed)
}

func (cfg *apiConfig) getFollowsHandleFunc(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey).(database.User)
	if !ok {
		respondWithError(w, 500, "Failed to retrieve user data from context")
		return
	}
	follows, err := cfg.DB.GetFollows(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	fs := []Follow{}
	for _, f := range follows {
		fs = append(fs, toFollow(f))
	}
	respondWithJSON(w, 200, fs)
}

func (cfg *apiConfig) unfollowHandleFunc(w http.ResponseWriter, r *http.Request) {
	i, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, 404, "Invalid ID: "+chi.URLParam(r, "id"))
		return
	}
	user, ok := r.Context().Value(userKey).(database.User)
	if !ok {
		respondWithError(w, 500, "Failed to retrieve user data from context")
		return
	}
	param := database.UnfollowParams{
		ID:     i,
		UserID: user.ID,
	}
	_, err = cfg.DB.Unfollow(r.Context(), param)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	w.WriteHeader(202)
}
