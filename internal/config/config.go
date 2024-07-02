package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort             string `mapstructure:"APP_PORT"`
	GetByPassportDomain string `mapstructure:"GETBYPASSPORTDOMAIN"`
	PostgresHost        string `mapstructure:"POSTGRES_HOST"`
	PostgresPort        string `mapstructure:"POSTGRES_PORT"`
	PostgresUser        string `mapstructure:"POSTGRES_USER"`
	PostgresPassword    string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDBName      string `mapstructure:"POSTGRES_DBNAME"`
}

func LoadConfig(path string) (c Config, err error) {

	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&c)

	return
}

func MustLoad(path string) Config {
	c, err := LoadConfig(path)
	if err != nil {
		log.Fatal(err)
	}
	return c
}
