package queue

import (
	"context"
	"warehouse-management-api/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockEmailProducer struct{ mock.Mock }

func (m *MockEmailProducer) PushEmailJob(ctx context.Context, job entity.EmailJob) error {
	return m.Called(ctx, job).Error(0)
}
