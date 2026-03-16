package handler

import (
	"net/http"
)

type ServiceInterface interface {
}

func Test(s ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
