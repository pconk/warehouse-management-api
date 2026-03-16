package handler

import (
	"github.com/stretchr/testify/mock"
)

type MockHealthRepo struct {
	mock.Mock
}

func (m *MockHealthRepo) Ping() error {
	args := m.Called()
	return args.Error(0)
}
