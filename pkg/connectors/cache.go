package connectors

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/dutt23/lms/config"
	"github.com/redis/go-redis/v9"
)

type CacheConnector interface {
	Connector
	DB(ctx context.Context) *redis.Client
}

type dragonFlyConnector struct {
	cfg        *config.CacheConfig
	Connection *redis.Client
}

func NewCacheConnector(config *config.CacheConfig) CacheConnector {
	return &dragonFlyConnector{cfg: config}
}

func (dragonFlyConn *dragonFlyConnector) connectionString() string {
	return fmt.Sprintf("%s:%d", dragonFlyConn.cfg.Host, dragonFlyConn.cfg.Port)
}

// provide a debug name for connector
func (dragonFlyConn *dragonFlyConnector) Name() string {
	return fmt.Sprintf("DRAGONFLY %s:%d", dragonFlyConn.cfg.Host, dragonFlyConn.cfg.Port)
}

func (dragonFlyConn *dragonFlyConnector) DB(ctx context.Context) *redis.Client {
	return dragonFlyConn.GetConnection()
}

// only connect the call usually made by main.go to create a connection with given configuration
// anyway can be called anywhere as config is will always be in socpe of connect
func (dragonFlyConn *dragonFlyConnector) Connect(ctx context.Context) error {
	opt := &redis.Options{
		Addr:     dragonFlyConn.connectionString(),
		PoolSize: dragonFlyConn.cfg.MaxConnection,
		Password: dragonFlyConn.cfg.Auth.Password,
		DB:       dragonFlyConn.cfg.Db,
	}
	if dragonFlyConn.cfg.InsecureSkipVerify {
		opt.TLSConfig = &tls.Config{
			InsecureSkipVerify: dragonFlyConn.cfg.InsecureSkipVerify,
		}
	}
	client := redis.NewClient(opt)

	dragonFlyConn.Connection = client
	fmt.Printf("Created new client for redis with name: %s", dragonFlyConn.Name())

	if ok := dragonFlyConn.IsConnected(ctx); !ok {
		fmt.Errorf("could not connect to redis client")
	}
	return nil
}

// getting connection to use if anyone wants to use the connection
func (dragonFlyConn *dragonFlyConnector) GetConnection() *redis.Client {
	return dragonFlyConn.Connection
}

// Return boolean status if connected or not
func (dragonFlyConn *dragonFlyConnector) IsConnected(ctx context.Context) bool {

	fmt.Printf("Pinging redis server.")
	pingResponse, err := dragonFlyConn.Connection.Ping(ctx).Result()
	if err != nil {
		fmt.Errorf("Error while pinging redis server. %v", err)
		return false
	}
	fmt.Printf("Return from ping command %v", pingResponse)
	return true
}

func (dragonFlyConn *dragonFlyConnector) Disconnect(ctx context.Context) error {
	fmt.Printf("Disconnecting with redis client.")

	err := dragonFlyConn.Connection.Close()
	if err != nil {
		fmt.Errorf("Failed to disconnect redis client. %v", err)
		return err
	}
	fmt.Printf("Disconnected successful redis client.")
	// anyway nil the connection reference
	dragonFlyConn.Connection = nil
	return err

}
