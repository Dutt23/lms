package config

type DBConfig struct {
	Host               string    `mapstructure:"host" validate:"required"`
	Port               int       `mapstructure:"port"`
	Auth               BasicAuth `mapstructure:"auth"`
	DBName             string    `mapstructure:"db_name" validate:"required"`
	MaxIdealConnection int       `mapstructure:"max_ideal_connection" validate:"required"`
	MaxOpenConnection  int       `mapstructure:"max_open_connection" validate:"required"`
	SslMode            string    `mapstructure:"ssl_mode" validate:"required"`
}