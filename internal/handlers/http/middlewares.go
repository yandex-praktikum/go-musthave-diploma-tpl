package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

type authClient interface {
	AuthUser(ctx context.Context, token domain.SessionToken) (domain.SessionInfo, error)
}

type AuthMiddleware struct {
	auth authClient
}

func NewAuthMiddleware(auth authClient) *AuthMiddleware {
	return &AuthMiddleware{
		auth: auth,
	}
}

func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(AuthorizationKey)

		if token == "" {
			c, err := r.Cookie(AuthorizationKey)
			if err != nil {
				switch {
				case errors.Is(err, http.ErrNoCookie):
					writeError(r.Context(), w,
						serviceerrors.NewUnauthorized().Wrap(err, "authorized cookie not found"))
					return
				}

				writeError(r.Context(), w, err)
				return
			}

			token = c.Value
		}

		session, err := am.auth.AuthUser(r.Context(), domain.SessionToken{
			Token: token,
		})
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				writeError(r.Context(), w,
					serviceerrors.NewUnauthorized().Wrap(err, "authorized cookie not found"))
			}
			writeError(r.Context(), w, err)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func getUserID(ctx context.Context) (domain.ID, bool) {
	val, ok := ctx.Value(userIDKey).(domain.ID)
	return val, ok
}

// AddLoggerToContextMiddleware помещает logger в context
func AddLoggerToContextMiddleware(sugarLogger *zap.SugaredLogger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ctx = logger.ToContext(ctx, sugarLogger)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

// RequestMiddleware middleware для логирования запросов
func RequestMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() {
				logger.Infof(r.Context(), "request: url: %s; method: %s; processing time: %s",
					r.URL.String(), r.Method, time.Since(start).String())
			}()

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// ResponseMiddleware middleware для логирования ответов
func ResponseMiddleware() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			updatedWriter := NewWriterWithLogging(w)
			defer func() {
				defer func() {
					logger.Infof(r.Context(), "response: status code: %d, datasize: %d bytes",
						updatedWriter.statusCode,
						updatedWriter.responseSize)
				}()
			}()

			next.ServeHTTP(updatedWriter, r)
		}

		return http.HandlerFunc(fn)
	}
}

// WriterWithLogging реализация интерфейса writer для перехвата информации ответа
type WriterWithLogging struct {
	statusCode   int
	responseSize int

	baseWriter http.ResponseWriter
}

// NewWriterWithLogging создание нового WriterWithLogging объекта
func NewWriterWithLogging(baseWriter http.ResponseWriter) *WriterWithLogging {
	return &WriterWithLogging{
		baseWriter: baseWriter,
	}
}

// Write ...
func (w *WriterWithLogging) Write(b []byte) (int, error) {
	w.responseSize = len(b)
	return w.baseWriter.Write(b)
}

// WriteHeader ...
func (w *WriterWithLogging) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.baseWriter.WriteHeader(statusCode)
}

// Header ...
func (w *WriterWithLogging) Header() http.Header {
	return w.baseWriter.Header()
}
