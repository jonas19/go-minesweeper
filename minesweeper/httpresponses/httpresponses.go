package httpresponses

import (
	"encoding/json"
	"net/http"
)

type responseStruct struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func SendJSONResponse(w http.ResponseWriter, status string, message string) {
	var statusCode int
	if status == "ok" {
		statusCode = http.StatusOK
	} else {
		statusCode = http.StatusBadRequest
	}

	response := responseStruct{
		Status:  status,
		Message: message,
	}

	json, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(json))
}

func SendTextResponse(w http.ResponseWriter, status string, message string) {
	var statusCode int
	if status == "ok" {
		statusCode = http.StatusOK
	} else {
		statusCode = http.StatusBadRequest
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
