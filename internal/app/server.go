package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/r4start/go-musthave-diploma-tpl/internal/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
)

const (
	CompressionLevel = 7
)

var (
	ErrBadContentType = errors.New("bad content type in request")
	ErrBodyUnmarshal  = errors.New("failed to unmarshal request body")
)

type userAuthRequest struct {
	Login    string
	Password string
}

type Server struct {
	*chi.Mux
	ctx         context.Context
	logger      *zap.Logger
	userStorage storage.UserStorage
}

func NewServer(ctx context.Context, logger *zap.Logger, userStorage storage.UserStorage) (*Server, error) {
	server := &Server{
		Mux:         chi.NewMux(),
		ctx:         ctx,
		logger:      logger,
		userStorage: userStorage,
	}

	server.Use(middleware.NoCache)
	server.Use(middleware.Compress(CompressionLevel))
	server.Use(DecompressGzip)

	server.Post("/api/user/register", server.apiUserRegister)

	server.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusBadRequest)
	})

	return server, nil
}

func (s *Server) apiUserRegister(w http.ResponseWriter, r *http.Request) {
	authData := userAuthRequest{}
	if err := s.apiParseRequest(r, &authData); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if err := s.userStorage.Add(&storage.UserAuthorization{
		UserName: authData.Login,
		Secret:   []byte(authData.Password),
	}); err != nil {
		if errors.Is(err, storage.ErrDuplicateUser) {
			http.Error(w, "", http.StatusConflict)
			return
		}
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) apiParseRequest(r *http.Request, body interface{}) error {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		s.logger.Error("bad content type", zap.String("content_type", contentType))
		return ErrBadContentType
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("failed to read request body", zap.Error(err))
		return err
	}

	if err = json.Unmarshal(b, &body); err != nil {
		s.logger.Error("failed to unmarshal request json", zap.Error(err))
		return ErrBodyUnmarshal
	}

	return nil
}
