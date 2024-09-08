package handlers

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/with0p/gophermart/internal/mock"
)

func setup(t *testing.T) (*gomock.Controller, *mock.MockService, *HandlerUserAPI) {
	ctrl := gomock.NewController(t)
	mockService := mock.NewMockService(ctrl)
	h := &HandlerUserAPI{service: mockService}
	return ctrl, mockService, h
}
