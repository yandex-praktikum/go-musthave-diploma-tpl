package register

import (
	"bytes"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/auth"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type returnBody struct {
	saveTableUser               error
	checkTableUserLogin         error
	checkTableUserPassword      bool
	saveTableUserAndUpdateToken error
	withIncorrectBody           bool
}

func TestPost(t *testing.T) {

	tests := []struct {
		name           string
		requestBody    RequestBody
		returnBody     returnBody
		expectedStatus int
	}{
		{
			name: "Successful registration and authentication",
			requestBody: RequestBody{
				Login:    "test",
				Password: "password",
			},
			returnBody: returnBody{
				checkTableUserPassword: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Invalid request body",
			requestBody: RequestBody{},
			returnBody: returnBody{
				withIncorrectBody: true,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "User already exists",
			requestBody: RequestBody{
				Login:    "existing",
				Password: "password",
			},
			returnBody: returnBody{
				checkTableUserLogin: customerrors.ErrUserAlreadyExists,
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Registration error",
			requestBody: RequestBody{
				Login:    "test",
				Password: "password",
			}, returnBody: returnBody{
				checkTableUserLogin: customerrors.ErrNotFound,
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Authentication error",
			requestBody: RequestBody{
				Login:    "test",
				Password: "password",
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create logger
			logger := logger.NewLogger(logger.WithLevel("info"))

			// Create mock service auth
			repo := auth.NewMockStorageAuth(ctrl)

			authService := auth.NewService([]byte("test"), []byte("test"), repo)
			passwordHash := authService.HashPassword(tt.requestBody.Password)

			repo.EXPECT().CheckTableUserLogin(tt.requestBody.Login).Return(tt.returnBody.checkTableUserLogin).AnyTimes()
			repo.EXPECT().SaveTableUser(tt.requestBody.Login, passwordHash).Return(tt.returnBody.saveTableUser).AnyTimes()
			repo.EXPECT().CheckTableUserPassword(tt.requestBody.Login).Return(passwordHash, tt.returnBody.checkTableUserPassword).AnyTimes()
			repo.EXPECT().SaveTableUserAndUpdateToken(tt.requestBody.Login, gomock.Any()).Return(tt.returnBody.saveTableUserAndUpdateToken).AnyTimes()

			// Create handlers
			handlers := NewHandlers(authService, logger)

			req, err := http.NewRequest("POST", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.returnBody.withIncorrectBody {
				req.Body = io.NopCloser(bytes.NewBufferString("incorrect body"))
			} else {
				req.Body = jsonReader(tt.requestBody)
			}

			w := httptest.NewRecorder()
			handlers.Post(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func jsonReader(v interface{}) io.ReadCloser {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return io.NopCloser(bytes.NewBuffer(b))
}
