package config

import (
	"github.com/spf13/viper"
)

type config struct {
	ServerHost string `mapstructure:"SERVER_HOST"`
	ServerPort string `mapstructure:"SERVER_PORT"`

	DBUser     string `mapstructure:"DB_USER"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBPort     string `mapstructure:"DB_PORT"`
	SSLMode    string `mapstructure:"SSL_MODE"`

	RedisHost string `mapstructure:"R_HOST"`
	RedisPort string `mapstructure:"R_PORT"`

	MongoHost string `mapstructure:"M_HOST"`
}

func New(filename, path string) (*config, error) {
	cfg := new(config)

	viper.AddConfigPath(path)
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
