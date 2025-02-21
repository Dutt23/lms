package config

import (
	"log"
	"os"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Name              string      `mapstructure:"service_name" validate:"required"`
	DBSource          string      `mapstructure:"db_source" validate:"required"`
	MigrationUrl      string      `mapstructure:"migration_url" validate:"required"`
	Version           string      `mapstructure:"version" validate:"required"`
	Host              string      `mapstructure:"host" validate:"required"`
	Secret            string      `mapstructure:"secret" validate:"required"`
	Port              int         `mapstructure:"port" validate:"required"`
	LogLevel          string      `mapstructure:"log_level" validate:"required"`
	DbConfig          DBConfig    `mapstructure:"db" validate:"required"`
	CacheConfig       CacheConfig `mapstructure:"cache" validate:"required"`
	TokenSymmetricKey string      `mapstructure:"token_symmetric_key" validate:"required"`
	QueuePort int                 `mapstructure:"queue_port" validate:"required"`
}

// reading config and intializing configs for application
func InitConfig() (*viper.Viper, error) {
	vConfig := viper.NewWithOptions(viper.KeyDelimiter("__"))

	vConfig.AddConfigPath(".")
	vConfig.SetConfigName(".env")
	path := os.Getenv("ENV_PATH")
	if path != "" {
		log.Printf("env path %v", path)
		vConfig.SetConfigFile(path)
	}
	vConfig.SetConfigType("env")
	vConfig.AutomaticEnv()
	err := vConfig.ReadInConfig()
	if err == nil {
		log.Printf("Error while reading the config")
	}

	//
	setDefault(vConfig)
	if err = vConfig.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		log.Printf("Reading from env varaibles.")
	}

	return vConfig, nil
}

func setDefault(v *viper.Viper) {
	v.SetDefault("SERVICE_NAME", "lms-project")
	v.SetDefault("VERSION", "0.0.1")
	v.SetDefault("HOST", "localhost")
	v.SetDefault("PORT", "")
	v.SetDefault("LOG_LEVEL", "debug")
	//

	v.SetDefault("DB__HOST", "")
	v.SetDefault("DB__PORT", "")
	v.SetDefault("DB__DB_NAME", "")
	v.SetDefault("DB__AUTH__USER", "")
	v.SetDefault("DB__AUTH__PASSWORD", "<>")
	v.SetDefault("DB__MAX_OPEN_CONNECTION", 10)
	v.SetDefault("DB__MAX_IDEAL_CONNECTION", 10)
	v.SetDefault("DB__SSL_MODE", "disable")
}

// Getting application config from viper
func GetApplicationConfig(v *viper.Viper) (*AppConfig, error) {
	var config AppConfig
	err := v.Unmarshal(&config)
	if err != nil {
		log.Printf("%+v\n", err)
		return nil, err
	}

	// valdating the app config
	validate := validator.New()
	err = validate.Struct(&config)
	if err != nil {
		log.Printf("%+v\n", err)
		return nil, err
	}
	return &config, nil
}
