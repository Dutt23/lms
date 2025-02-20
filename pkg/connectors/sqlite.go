package connectors

import (
	"context"

	"gorm.io/gorm"
)

type SqliteConnector interface {
	Connector
	Query(ctx context.Context, qry string, dest interface{}) error
	DB(ctx context.Context) *gorm.DB
}