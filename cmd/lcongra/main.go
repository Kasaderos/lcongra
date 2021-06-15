package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/kasaderos/lcongra/config"
	exchange "github.com/kasaderos/lcongra/exchange/fake"
	"github.com/kasaderos/lcongra/service"
)

func main() {
	conf := config.ReadConfig("template")
	log.Printf("%+v", conf)
	// logFile, err := os.OpenFile("service.log", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	logger := log.New(os.Stdout, "[obs-binance] ", log.Flags())
	binance := exchange.NewExchange(logger)

	chanMsg := make(chan string)

	// logger = log.New(logFile, "[reporter] ", log.Flags())
	// reporter := service.NewReporter(conf.ClientURL, logger, chanMsg)

	ctx, quit := context.WithCancel(context.Background())

	logger = log.New(os.Stdout, "[observer] ", log.Flags())
	observer := service.NewObserver(binance, logger, chanMsg, []string{conf.Pair})
	go observer.Observe(ctx)

	logger = log.New(os.Stdout, "[mas] ", log.Flags())
	srv := service.NewAgentService(observer, nil, logger)

	handler := service.NewAgentServiceHandler(srv, ctx)

	_ = quit // TODO
	log.Println("service started 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
