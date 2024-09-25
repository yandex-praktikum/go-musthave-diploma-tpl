package authorize

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/auth"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type resultBody struct {
	saveTableUser               error
	checkTableUserLogin         error
	checkTableUserPassword      bool
	saveTableUserAndUpdateToken error
	withIncorrectBody           bool
	testErrors                  error
}

func TestHandlerPost(t *testing.T) {

	tests := []struct {
		name           string
		requestBody    RequestBody
		resultBody     resultBody
		expectedStatus int
	}{
		{
			name: "Successful login",
			resultBody: resultBody{
				checkTableUserPassword: true,
			},
			expectedStatus: http.StatusOK,
		},

		{
			name: "invalid request body",
			resultBody: resultBody{
				withIncorrectBody: true,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "not found authorization",
			resultBody: resultBody{
				saveTableUserAndUpdateToken: customerrors.ErrNotFound,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "incorrect password",
			resultBody: resultBody{
				checkTableUserPassword:      true,
				saveTableUserAndUpdateToken: customerrors.ErrIsTruePassword,
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "can't authorize user",
			resultBody: resultBody{
				checkTableUserPassword:      true,
				saveTableUserAndUpdateToken: errors.New("can't authorize user"),
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()

			loger := logger.NewLogger()

			ctrl := gomock.NewController(t)

			repo := auth.NewMockStorageAuth(ctrl)

			authService := auth.NewService([]byte("test"), []byte("test"), repo)

			passwordHash := authService.HashPassword(tt.requestBody.Password)
			repo.EXPECT().CheckTableUserLogin(ctx, tt.requestBody.Login).Return(tt.resultBody.checkTableUserLogin).AnyTimes()
			repo.EXPECT().CheckTableUserPassword(ctx, tt.requestBody.Password).Return(passwordHash, tt.resultBody.checkTableUserPassword).AnyTimes()
			repo.EXPECT().SaveTableUser(tt.requestBody.Login, gomock.Any()).Return(tt.resultBody.saveTableUser).AnyTimes()
			repo.EXPECT().SaveTableUserAndUpdateToken(tt.requestBody.Login, gomock.Any()).Return(tt.resultBody.saveTableUserAndUpdateToken).AnyTimes()
			handler := NewHandler(ctx, authService, loger)

			req, err := http.NewRequest("POST", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.resultBody.withIncorrectBody {
				req.Body = io.NopCloser(bytes.NewBufferString("incorect body"))
			} else {
				req.Body = jsonReader(tt.requestBody)
			}

			w := httptest.NewRecorder()
			handler.Post(w, req)

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
