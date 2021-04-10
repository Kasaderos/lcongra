package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kasaderos/lcongra/config"
	"github.com/kasaderos/lcongra/exchange/binance"
	"github.com/kasaderos/lcongra/service"
)

func main() {
	conf := config.ReadConfig("template")
	logFile, err := os.OpenFile("../../service.log", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(logFile, "[binance] ", log.Flags())
	binance := binance.NewExchange(logger)

	chanMsg := make(chan string)
	observer := service.NewObserver(binance, logger, chanMsg, []string{conf.Pair})

	logger = log.New(logFile, "[reporter] ", log.Flags())
	reporter := service.NewReporter(conf.ClientURL, logger, chanMsg)

	ctx, quit := context.WithCancel(context.Background())
	go observer.Observe(ctx)
	go reporter.Report(ctx)

	logger = log.New(logFile, "[bot] ", log.Flags())
	queue := service.NewOrderQueue()
	bot := service.NewBot(queue, binance, logger)
	go bot.StartSM(ctx)

	userChan := make(chan string)
	logger = log.New(logFile, "[handler] ", log.Flags())
	handler := NewHttpHandler(queue, userChan, logger)
	go http.ListenAndServe(":8080", handler)

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

type HttpHandler struct {
	queue    *service.OrderQueue
	userChan chan string
	logger   *log.Logger
}

func NewHttpHandler(queue *service.OrderQueue, userChan chan string, logger *log.Logger) http.Handler {
	return &HttpHandler{
		queue:    queue,
		userChan: userChan,
		logger:   logger,
	}
}

type Message struct {
	Type string
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
	case "POST":
		data, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Println(err)
			return
		}
		var msg Message
		err = json.Unmarshal(data, &msg)
		if err != nil {
			h.logger.Println(err)
			return
		}
		// h.userChan <- string(msg)
		w.WriteHeader(http.StatusAccepted)
	}
}
