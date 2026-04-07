package errs

import (
	"encoding/json"
	"net/http"
)

type problemDetail struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func Write(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	body, _ := json.Marshal(problemDetail{
		Title:  http.StatusText(status),
		Status: status,
		Detail: msg,
	})
	_, _ = w.Write(body)
}
