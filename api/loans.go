package api

import (
	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/pkg/connectors"
)


type loansApi struct {
  config *config.AppConfig
	db     connectors.SqliteConnector
}

func NewLoansApi(config *config.AppConfig, db connectors.SqliteConnector) *loansApi {
	return &loansApi{
		config: config,
		db:     db,
	}
}