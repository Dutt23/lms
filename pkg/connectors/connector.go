package connectors

import "context"

type Connector interface {
	Connect(ctx context.Context) error
	Name() string
	IsConnected(ctx context.Context) bool
	Disconnect(ctx context.Context) error
}