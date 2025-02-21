package workers

import (
	"context"

	"github.com/hibiken/asynq"
)

type Proccessor interface {
	Start() error
	Process(ctx context.Context, task *asynq.Task) error
}
