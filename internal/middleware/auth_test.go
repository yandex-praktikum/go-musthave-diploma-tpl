package middleware

import (
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/auth"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewareValidAuth(t *testing.T) {
	tests := []struct {
		name           string
		accessToken    string
		expentedStatus int
	}{
		{
			name:           "No access token",
			accessToken:    "",
			expentedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid access token",
			accessToken:    "invalid_token",
			expentedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid Access token",
			accessToken:    "valid_token",
			expentedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := auth.NewMockStorageAuth(ctrl)
			repo.EXPECT().SaveTableUserAndUpdateToken(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			authService := auth.NewService([]byte("test"), []byte("test"), repo)

			middlewareAuth := NewAuthMiddleware(authService)

			req, err := http.NewRequest("GET", "/api/user", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.accessToken != "" && tt.accessToken != "valid_token" {
				req.AddCookie(&http.Cookie{Name: string(AccessTokenKey), Value: tt.accessToken})
			}

			if tt.accessToken == "valid_token" {
				tokenTest, err := authService.GeneratedTokens("test")
				if err != nil {
					t.Fatal(err)
				}

				req.AddCookie(&http.Cookie{Name: string(AccessTokenKey), Value: tokenTest.AccessToken})
			}
			w := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			middlewareHandler := middlewareAuth.ValidAuth(handler)
			middlewareHandler.ServeHTTP(w, req)

			if w.Code != tt.expentedStatus {
				t.Errorf("expented code = %d, got = %d", tt.expentedStatus, w.Code)
			}
		})
	}
}
