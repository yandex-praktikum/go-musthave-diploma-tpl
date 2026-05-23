package server

import (
	"net/http"
	"runtime/debug"

	"go-musthave-diploma-tpl/internal/logger"
)

// Recovery middleware для восстановления после паник
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorWithStack("Panic recovered", nil,
					"error", err,
					"stack", string(debug.Stack()),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
