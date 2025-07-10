package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// JSONResponse is a standard format for all API responses
type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON writes a success response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	res := JSONResponse{
		Success: true,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		logrus.Error("Error encoding success response: ", err)
	}
}

// Error writes an error response
func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	res := JSONResponse{
		Success: false,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		logrus.Error("Error encoding error response: ", err)
	}
}


      

