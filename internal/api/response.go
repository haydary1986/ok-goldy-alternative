package api

import (
	"encoding/json"
	"net/http"
)

// Envelope is the consistent response shape for all JSON API responses.
type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
	Meta    *PageMeta  `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PageMeta struct {
	Total     int    `json:"total,omitempty"`
	Page      int    `json:"page,omitempty"`
	PageSize  int    `json:"page_size,omitempty"`
	NextPage  string `json:"next_page_token,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func WriteData(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, Envelope{Success: true, Data: data})
}

func WritePage(w http.ResponseWriter, status int, data any, meta PageMeta) {
	WriteJSON(w, status, Envelope{Success: true, Data: data, Meta: &meta})
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message},
	})
}
