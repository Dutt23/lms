package workers

import (
	"context"
	"fmt"

	"github.com/dutt23/lms/config"
	"github.com/hibiken/asynq"
)

const (
	CriticalQueue       = "critical"
	defaultQueue        = "default"
	taskOrdersAnalytics = "task:orders_analytics"
)

type analyticsTaskProcessor struct {
	server *asynq.Server
}

func NewAnalyticsTaskProcessor(config *config.AppConfig) Proccessor {
	redisOpts := asynq.RedisClientOpt{
		Addr: fmt.Sprintf("%s:%s", config.CacheConfig.Host, config.CacheConfig.Port),
	}
	server := asynq.NewServer(redisOpts, asynq.Config{
		Queues: map[string]int{
			CriticalQueue: 10,
			defaultQueue:  3,
			"low":         1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			fmt.Println("Task processing has failed with error ", err)
		}),
		Logger: NewLogger(),
	})

	return &analyticsTaskProcessor{
		server,
	}
}

func (processor *analyticsTaskProcessor) Process(ctx context.Context, task *asynq.Task) error {
	return nil
}

func (processor *analyticsTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(taskOrdersAnalytics, processor.Process)
	return processor.server.Start(mux)
}
