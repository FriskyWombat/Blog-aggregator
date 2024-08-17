package main

import (
	"encoding/json"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type error struct {
		Error string `json:"error"`
	}
	e := error{Error: msg}
	respondWithJSON(w, code, e)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func healthzHandleFunc(w http.ResponseWriter, r *http.Request) {
	type statusResp struct {
		Status string `json:"status"`
	}
	s := statusResp{"ok"}
	respondWithJSON(w, 200, s)
}

func errorHandleFunc(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}
