package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kasaderos/lcongra/config"
	exchange "github.com/kasaderos/lcongra/exchange/fake"
	"github.com/kasaderos/lcongra/service"
)

func main() {
	// conf := config.ReadConfig("template")
	_ = config.ReadConfig("template")
	logFile, err := os.OpenFile("../../service.log", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// logger := log.New(logFile, "[binance] ", log.Flags())
	// binance := binance.NewExchange(logger, conf.ApiKey, conf.ApiSecret)
	logger := log.New(logFile, "[binance] ", log.Flags())
	binance := exchange.NewExchange(logger)

	// chanMsg := make(chan string)

	logger = log.New(logFile, "[reporter] ", log.Flags())
	// reporter := service.NewReporter(conf.ClientURL, logger, chanMsg)

	ctx, quit := context.WithCancel(context.Background())

	// observer := service.NewObserver(binance, logger, chanMsg, []string{conf.Pair})
	// go observer.Observe(ctx)
	// go reporter.Report(ctx)

	logger = log.New(logFile, "[bot] ", log.Flags())
	queue := service.NewOrderQueue()
	bot := service.NewBot(queue, binance, logger)
	go bot.StartSM(ctx)

	userChan := make(chan string)
	logger = log.New(logFile, "[handler] ", log.Flags())

	go service.Autotrade(
		ctx,
		"BTC-USDT",
		"3m",
		queue,
		binance,
	)
	log.Println("service started 8080")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	select {
	case msg := <-userChan:
		if msg == "STOPALL" {
			quit()
		}
	case <-sig:
		quit()
	}
}
