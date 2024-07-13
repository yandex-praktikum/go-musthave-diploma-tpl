package client

import (
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
)

// MockHTTPClient is a mocks implementation of httpClient
type MockHTTPClient struct {
	mock.Mock
}

// NewRequest is a method on MockHTTPClient that satisfies ClientHTTP interface
func (m *MockHTTPClient) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	args := m.Called(method, url, body)
	return args.Get(0).(*http.Request), args.Error(1)
}

// Do is a method on MockHTTPClient that satisfies ClientHTTP interface
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockResponse is a mocks implementation of http.Response
type MockResponse struct {
	mock.Mock
}

// Body is a method on MockResponse that satisfies io.ReadCloser interface
func (m *MockResponse) Body() io.ReadCloser {
	return m.Called().Get(0).(io.ReadCloser)
}

// Read is a method on MockResponse that satisfies io.ReadCloser interface
func (m *MockResponse) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

// Close is a method on MockResponse that satisfies io.ReadCloser interface
func (m *MockResponse) Close() error {
	return m.Called().Error(0)
}

type BodyMock struct {
}

func (b BodyMock) Read(p []byte) (n int, err error) {
	return 0, err
}

func (b BodyMock) Close() error {
	return nil
}
