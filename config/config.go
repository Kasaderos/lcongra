package config

import (
	"log"

	"fmt"
	"os"
	"github.com/spf13/viper"
)

type Configuration struct {
	ClientURL string
	Interval  string
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
	viper.AddConfigPath(fmt.Sprintf("%s/config", os.Getenv("APPDIR")))
	config := new(Configuration)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	config.ClientURL = viper.GetString("client-url")
	config.LogFile = viper.GetString("log-file")
	config.QueueSize = viper.GetInt("order-queue-size")
	config.Pair = viper.GetString("pair")
	config.Interval = viper.GetString("interval")
	config.ApiKey = ""
	config.ApiSecret = ""

	return config
}
