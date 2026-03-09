package response

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-Id"

type Envelope struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
}

func WriteData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: data})
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	requestID := w.Header().Get(requestIDHeader)
	if requestID == "" {
		requestID = uuid.NewString()
		w.Header().Set(requestIDHeader, requestID)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Error: &APIError{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	})
}
