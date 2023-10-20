package handler

// import (
// 	"bytes"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/service"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/require"
// )

// func TestHandler_SingUp(t *testing.T) {

// 	login := service.RandStr(10)
// 	password := service.RandStr(12)
// 	m := []byte(`
//         {
//             "login": "` + login + `",
//             "password": "` + password + `"
//         }
//     `)
// 	var testTableValue = []struct {
// 		name   string
// 		body   []byte
// 		status int
// 	}{

// 		"test autorisation",
// 		{m, http.StatusNotFound},
// 		// {"/value/gauge1/testSetGet40", "404 page not found", http.StatusNotFound},
// 	}

// 	handler := http.HandlerFunc(SingUp)
// 	srv := httptest.NewServer(handler)
// 	defer srv.Close()

// 	for _, v := range testTableValue {
// 		t.Run(v.name, func(t *testing.T) {
// 			w := httptest.NewRecorder()
// 			c, r := gin.CreateTestContext(w)
// 			r.POST("/", handler)
// 			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(v.body))

// 			r.ServeHTTP(w, c.Request)

// 			result := w.Result()

// 			defer result.Body.Close()
// 			require.Equal(t, v.status, result.StatusCode)
// 			// require.Equal(t, v.want, result.Header["Content-Type"])

// 		})
// 	}

// }
