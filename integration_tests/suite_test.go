package integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brisk84/gofemart/domain"
	"github.com/brisk84/gofemart/internal/app"
	"github.com/brisk84/gofemart/internal/config"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	app    *app.App
	server *httptest.Server
}

func (s *TestSuite) SetupSuite() {
	var err error
	s.app, err = app.New(config.Config{})
	s.Require().NoError(err)

	s.server = httptest.NewServer(s.app.HTTPServer.Handler)
}

func (s *TestSuite) TearDownSuite() {
	s.server.Close()
}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestSuite_CreateUser() {

	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("Name %d", i)
		user := domain.User{
			Name: name,
		}

		body := bytes.Buffer{}
		json.NewEncoder(&body).Encode(user)

		request, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/v1/users", &body)
		s.Require().NoError(err)

		response, err := http.DefaultClient.Do(request)
		s.Require().NoError(err)

		s.Require().Equal(http.StatusOK, response.StatusCode)

		err = json.NewDecoder(response.Body).Decode(&user)
		s.Require().NoError(err)

		s.Require().Equal(name, user.Name)
		s.Require().Equal(int64(i), user.ID)
	}
}
