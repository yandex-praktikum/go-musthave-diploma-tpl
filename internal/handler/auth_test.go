// package handler

// import (
// 	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/service"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/require"
// )

// func testRequest(t *testing.T, ts *gin.Engine, path string, b []byte) *http.Response {
// 	req, err := http.NewRequest("POST", "/value/counter1/", bytes.NewReader(b))
// 	require.NoError(t, err)

// 	rr := httptest.NewRecorder()
// 	ts.ServeHTTP(rr, req)
// 	require.NoError(t, err)

// 	resp := rr.Result()

// 	respBody, err := io.ReadAll(resp.Body)
// 	require.NoError(t, err)

// 	fmt.Println("respBody", string(respBody))
// 	return resp
// }

// func TestHandler_SingUp(t *testing.T) {

// 	router := gin.Default()

// 	login := service.RandStr(10)
// 	password := service.RandStr(12)
// 	m := []byte(`
//         {
//             "login": "` + login + `",
//             "password": "` + password + `"
//         }
//     `)
// 	var testTableValue = []struct {
// 		body   []byte
// 		status int
// 	}{

// 		// проверим на ошибочный запрос
// 		{m, http.StatusNotFound},
// 		// {"/value/gauge1/testSetGet40", "404 page not found", http.StatusNotFound},
// 	}
//     cfg := &config.ConfigServer{Port: "localhost:8080"}
//     h := NewHandler(stor, cfg, logger.Logger{})

// 	for _, v := range testTableValue {
// 		t.Run(tt.name, func(t *testing.T) {
// 			w := httptest.NewRecorder()
// 			c, r := gin.CreateTestContext(w)
// 			r.POST("/", h.)
// 			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

// 			r.ServeHTTP(w, c.Request)

// 			result := w.Result()

// 			defer result.Body.Close()
// 			require.Equal(t, tt.status, result.StatusCode)
// 			require.Equal(t, tt.want, result.Header["Content-Type"])

// 		})
// 	}

// }
