package workers

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeBooksAnalyticsPayload(ctx context.Context, payload *BookAnalyticsPayload, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)

	return &RedisTaskDistributor{
		client: client,
	}
}
