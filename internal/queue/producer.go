package queue

import (
	"context"
	"encoding/json"
	"warehouse-management-api/internal/entity"

	"github.com/redis/go-redis/v9"
)

type EmailProducer struct {
	redis     *redis.Client
	queueName string
}

type EmailProducerInterface interface {
	PushEmailJob(ctx context.Context, job entity.EmailJob) error
}

func NewEmailProducer(r *redis.Client, qname string) EmailProducerInterface {
	return &EmailProducer{redis: r, queueName: qname}
}

func (p *EmailProducer) PushEmailJob(ctx context.Context, job entity.EmailJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	// "email_queue" harus sama dengan yang ada di .env worker
	return p.redis.RPush(ctx, p.queueName, data).Err()
}
