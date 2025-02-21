package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	"github.com/hibiken/asynq"
)

const (
	CriticalQueue       = "critical"
	defaultQueue        = "default"
	taskOrdersAnalytics = "task:orders_analytics"
)

type analyticsTaskProcessor struct {
	server *asynq.Server
	cache  connectors.CacheConnector
}

type BookAnalyticsPayload struct {
	Book   *model.Book     `json:"book"`
	Loan   *model.BookLoan `json:"loan"`
	Member *model.Member   `json:"member"`
}

func (distributor RedisTaskDistributor) DistributeBooksAnalyticsPayload(ctx context.Context, payload *BookAnalyticsPayload, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("failed to marshal task payload for sending email")
	}
	task := asynq.NewTask(taskOrdersAnalytics, jsonPayload, opts...)
	taskInfo, err := distributor.client.EnqueueContext(ctx, task)

	if err != nil {
		return fmt.Errorf("failed to enqueue analytics task")
	}

	fmt.Println("Sent books analytical task with payload  ", taskInfo.Payload)
	return nil
}

func NewAnalyticsTaskProcessor(config *config.AppConfig, cache connectors.CacheConnector) Proccessor {
	redisOpts := asynq.RedisClientOpt{
		Addr: "0.0.0.0:6379",
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
		cache,
	}
}

func (processor *analyticsTaskProcessor) Process(ctx context.Context, task *asynq.Task) error {
	var payload BookAnalyticsPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		fmt.Println("Error in queue ", err)
		return fmt.Errorf("unable to un-marshal json for task %w", asynq.SkipRetry)
	}
	go processor.updateBookMonthCount(ctx, *payload.Book, *payload.Loan)
	go processor.updateAuthorWeeklyCount(ctx, *payload.Book, *payload.Loan)
	return nil
}

func (processor *analyticsTaskProcessor) updateBookMonthCount(ctx context.Context, book model.Book, loan model.BookLoan) error {
	cache := processor.cache.DB(ctx)
	key := fmt.Sprintf("SET_INTERNAL_ANALYTICS_BOOK_%d", book.Id)
	m := fmt.Sprintf("%d/%d", loan.LoanDate.Year(), loan.LoanDate.Month())
	cache.ZIncrBy(ctx, key, 1, m)
	return nil
}

func (processor *analyticsTaskProcessor) updateAuthorWeeklyCount(ctx context.Context, book model.Book, loan model.BookLoan) error {
	cache := processor.cache.DB(ctx)
	key := fmt.Sprintf("SET_INTERNAL_ANALYTICS_MEMBER_%d", loan.MemberId)
	year, week := loan.LoanDate.ISOWeek()
	m := fmt.Sprintf("%d/%d/%d", year, loan.LoanDate.Month(), week)
	cache.ZIncrBy(ctx, key, 1, m)
	return nil
}

func (processor *analyticsTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(taskOrdersAnalytics, processor.Process)
	return processor.server.Start(mux)
}
