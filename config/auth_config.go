package config

type BasicAuth struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}
