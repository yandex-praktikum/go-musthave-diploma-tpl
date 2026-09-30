package handler

import (
	"compress/gzip"

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
)

// Routes builds and returns the chi router with all gophermart endpoints registered.
func (h *apiHandler) Routes() *chi.Mux {
	router := chi.NewRouter()
	compressor := chimw.NewCompressor(gzip.DefaultCompression)
	router.Use(compressor.Handler)
	// Public endpoints
	router.Route("/api/user", func(r chi.Router) {
		r.With(requireJSONContent).Post("/register", h.registerUser)
		r.With(requireJSONContent).Post("/login", h.loginUser)

		// Private endpoints
		r.Group(func(r chi.Router) {
			r.Use(h.validateToken)

			r.Post("/orders", h.loadOrder)
			r.Get("/orders", h.listOrders)

			r.Get("/balance", h.getBalance)
			r.With(requireJSONContent).Post("/balance/withdraw", h.withdraw)
			r.Get("/withdrawals", h.listWithdrawals)
		})
	})

	return router
}
