package main

import (
	"log"
	"os"

	"github.com/kasaderos/lcongra/pkg/service"
)

func main() {
	logger := log.New(os.Stdout, "[reporter] ", log.Flags())
	chanMsg := make(chan string)
	reporter := service.NewReporter("", logger, chanMsg)
	_ = reporter
}
