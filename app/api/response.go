package api

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Error string `json:"error"`
}

func OKResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{Error: message})
}
