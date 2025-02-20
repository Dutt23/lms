package connectors

import "context"

type Connector interface {
	Connect(ctx context.Context) error
	Name() string
	Disconnect(ctx context.Context) error
}
