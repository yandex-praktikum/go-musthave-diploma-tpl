package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithLogging(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Testing logger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			middlewareAuth := WithLogging(handler)
			middlewareAuth.ServeHTTP(w, req)
		})
	}
}
