package servers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type httpServer struct {
	engine *chi.Mux
}

type ServiceOption func(s *httpServer)

var _ Server = &httpServer{}

func NewServer(opts ...ServiceOption) Server {
	s := &httpServer{
		engine: chi.NewRouter(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *httpServer) AddHandler(t RequestType, pattern string, handlerFn http.HandlerFunc) {
	switch t {
	case Post:
		s.engine.Post(pattern, handlerFn)
	case Get:
		s.engine.Get(pattern, handlerFn)
	}
}

func (s *httpServer) Start(addr string) error {
	if err := http.ListenAndServe(addr, s.engine); err != nil {
		return err
	}
	return nil
}
