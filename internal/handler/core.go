package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nikitaw13/go-musthave-diploma/internal/service"
	"go.uber.org/zap"
)

const queryTimeout = 3 * time.Second

// apiHandler serves the gophermart HTTP API on top of a LoyaltyService.
type apiHandler struct {
	svc    *service.LoyaltyService
	logger *zap.Logger
}

// NewAPIHandler creates a new API handler for the given loyalty service and logger.
func NewAPIHandler(svc *service.LoyaltyService, logger *zap.Logger) *apiHandler {
	return &apiHandler{
		svc:    svc,
		logger: logger,
	}
}

// decodeJSON reads the request body into T and writes an HTTP error response on failure;
// the bool reports success.
func decodeJSON[T any](w http.ResponseWriter, req *http.Request, logger *zap.Logger) (T, bool) {
	var v T
	var buf bytes.Buffer
	req.Body = http.MaxBytesReader(w, req.Body, maxRequestBodySize)

	_, err := buf.ReadFrom(req.Body)
	if maxBytesError, ok := errors.AsType[*http.MaxBytesError](err); ok {
		logger.Debug("content too large", zap.Error(maxBytesError))
		http.Error(w, "Content Too Large", http.StatusRequestEntityTooLarge)
		return v, false
	}
	if err != nil {
		logger.Error("error reading buffer", zap.Error(err))
		http.Error(w, "The request contains bad syntax or cannot be fulfilled", http.StatusBadRequest)
		return v, false
	}

	err = json.Unmarshal(buf.Bytes(), &v)
	if syntaxErr, ok := errors.AsType[*json.SyntaxError](err); ok {
		logger.Error("invalid JSON syntax", zap.Error(err))
		http.Error(w, fmt.Sprintf("invalid JSON syntax at offset %d", syntaxErr.Offset), http.StatusBadRequest)
		return v, false
	}
	if unmarshalTypeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		logger.Error("invalid type for field", zap.Error(err))
		http.Error(w, fmt.Sprintf("invalid type %q for field %q", unmarshalTypeErr.Value, unmarshalTypeErr.Field), http.StatusBadRequest)
		return v, false
	}

	if err != nil {
		logger.Error("error unmarshaling", zap.Error(err))
		http.Error(w, "The request contains bad syntax or cannot be fulfilled", http.StatusBadRequest)
		return v, false
	}
	return v, true
}
