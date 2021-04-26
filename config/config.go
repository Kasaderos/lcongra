package config

import (
	"log"

	"github.com/spf13/viper"
)

type Configuration struct {
	ClientURL string
	LogFile   string
	QueueSize int
	Pair      string
	ApiKey    string
	ApiSecret string
}

type ApiKeys struct {
	ApiKey    string `json:"apikey"`
	ApiSecret string `json:"apisecret"`
}

func ReadConfig(confname string) *Configuration {
	viper.SetConfigName("template")
	viper.AddConfigPath(".")
	config := new(Configuration)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	config.ClientURL = viper.GetString("client-url")
	config.LogFile = viper.GetString("log-file")
	config.QueueSize = viper.GetInt("order-queue-size")
	config.Pair = viper.GetString("pair")
	config.ApiKey = ""
	config.ApiSecret = ""

	return config
}
