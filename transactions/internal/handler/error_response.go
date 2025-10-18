package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nimbo1999/financeiro/transactions/internal/service"
)

type APIErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func ErrorResponse(code, message, details string) *APIErrorResponse {
	return &APIErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, apiError *APIErrorResponse) {
	w.WriteHeader(statusCode)
	// Ignoring error handling for brevity
	_ = json.NewEncoder(w).Encode(apiError)
}

func mapServiceError(err error) (int, *APIErrorResponse) {
	switch err {
	case service.ErrInvalidTransaction:
		return http.StatusBadRequest, ErrorResponse("INVALID_TRANSACTION", "Transaction is invalid", "Make sure to provide the user ID")
	default:
		// For unknown errors, log them but don't expose details
		log.Printf("Unmapped service error: %v", err)
		return http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", "An error occurred", "")
	}
}
