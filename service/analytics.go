package service

import (
	"context"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kasaderos/lcongra/exchange"
)

type Direction int

const (
	Stay Direction = 0
	Up   Direction = 1
	Down Direction = -1
)
const (
	app = "Rscript"
)

type Signal struct {
	Dir  Direction
	Time time.Time
}

func getDirection(pair string, interval string) Direction {
	// TODO
	cmd := exec.Command(app, "--vanilla", "../../scripts/la1_rf.R", interval, pair)

	output, err := cmd.Output()
	res := string(output)
	log.Println(res)
	if err != nil {
		log.Println("os exec output", err)
		return Stay
	}
	dir := strings.Split(res, " ")
	if len(dir) == 0 {
		return Stay
	}
	switch dir[0] {
	case "-1":
		return Down
	case "1":
		return Up
	case "0":
		return Stay
	}
	return Stay
}

func Autotrade(
	ctx context.Context,
	pair, interval string,
	ex exchange.Exchanger,
	signalChannel chan Signal,
) {
	logger := log.New(os.Stdout, "[autotrade] ", log.Default().Flags())

	var sleepDuration time.Duration
	switch interval {
	case "3m":
		sleepDuration = time.Minute * 3
	case "1m":
		sleepDuration = time.Minute
	}

	pairFormatted := ex.PairFormat(context.Background(), pair)
	logger.Println("started")
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		dir := getDirection(pairFormatted, interval)
		select {
		case signalChannel <- Signal{dir, time.Now()}:
		default:

		}
		time.Sleep(sleepDuration)
	}
}

func round(f float64, n int) float64 {
	base := math.Pow10(n)
	return math.Round(f*base) / base
}

/*
	or
*/
