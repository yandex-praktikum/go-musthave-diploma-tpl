package handler

import (
	"encoding/json"
	"net/http"
)

type NilType interface{}

func parseRequest[T any](r *http.Request) (T, error) {
	var data T
	err := json.NewDecoder(r.Body).Decode(&data)
	return data, err
}

func sendResponse[T any](w http.ResponseWriter, data T, err error, statusCode int) {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	if statusCode == http.StatusNoContent {
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		panic(err)
	}
}
