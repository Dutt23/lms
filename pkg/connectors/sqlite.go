package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/dutt23/lms/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type sqliteConnector struct {
	cfg *config.DBConfig
	db  *gorm.DB
}
type SqliteConnector interface {
	Connector
	DB(ctx context.Context) *gorm.DB
}

func (sql *sqliteConnector) DB(ctx context.Context) *gorm.DB {
	return sql.db.WithContext(ctx)
}

func NewSqliteConnector(config *config.DBConfig) SqliteConnector {
	return &sqliteConnector{cfg: config}
}

func (sql *sqliteConnector) Connect(ctx context.Context) error {
	db, err := gorm.Open(sqlite.Open("lms.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Errorf("Failed to open sqlite connection %s.", err)
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Errorf("Failed to create sqlite client connection pool %s.", err)
		return err
	}

	sqlDB.SetMaxIdleConns(sql.cfg.MaxIdealConnection)
	sqlDB.SetMaxOpenConns(sql.cfg.MaxOpenConnection)
	sqlDB.SetConnMaxLifetime(time.Hour)

	sql.db = db
	return nil
}

func (sql *sqliteConnector) Name() string {
	return fmt.Sprintf("SQLITE sql://%s:%d", sql.cfg.Host, sql.cfg.Port)
}

func (sql *sqliteConnector) Disconnect(ctx context.Context) error {
	fmt.Print("Disconnecting with postgres client.")
	db, err := sql.db.DB()
	if err != nil {
		fmt.Errorf("disconnecting with postgres client %s.", err)
		return err
	}
	err = db.Close()
	if err != nil {
		fmt.Printf("Disconnecting with postgres client %s.", err)
		return err
	}
	sql.db = nil
	return nil
}
