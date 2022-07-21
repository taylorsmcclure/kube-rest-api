package errors

import (
	// internal packages
	"github.com/taylorsmcclure/kube-server/internal/logger"
)

// Function to catch non-fatal errors/panics
func NonFatal() {
	if err := recover(); err != nil {
		logger.Log.Errorf("Non-fatal error occurred: %v", err)
	}
}

// Generic error struct to format a JSON response
type GenericError struct {
	Code    int    `json:"http_response_code"`
	Message string `json:"message"`
}
