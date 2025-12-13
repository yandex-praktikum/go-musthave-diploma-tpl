package handlers

import (
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (h *UserHandler) readRequestBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer h.closeBody(r.Body)
	return body, nil
}

func (h *UserHandler) closeBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		h.logger.Error("error close Body", zap.Error(err))
	}
}

func (h *UserHandler) handleConflictError(w http.ResponseWriter, err error) bool {
	/*
		if h.storage.IsDuplicateError(err) {
			http.Error(w, err.Error(), http.StatusConflict)
			return true
		}

	*/
	return false
}
