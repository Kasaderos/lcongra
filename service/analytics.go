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

func getDirection(pair string, interval string) Direction {
	app := "Rscript"
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
	exCtx context.Context,
	queue *OrderQueue,
	ex exchange.Exchanger,
) {
	logger := log.New(os.Stdout, "[autotrade] ", log.Default().Flags())

	var sleepDuration time.Duration
	switch interval {
	case "3m":
		sleepDuration = time.Minute * 3
	case "1m":
		sleepDuration = time.Second * 20
	}

	info, err := ex.GetInformation(exCtx, pair)
	if err != nil {
		logger.Println("[autotrade] ", err)
		return
	}
	pairFormatted := ex.PairFormat(context.Background(), pair)
	minAmount := 11.0
	fixedAmount := minAmount
	logger.Println("started")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// amount = max(balance/4, 11)
			if fixedAmount/4 > minAmount {
				fixedAmount /= 4
			}
			if !queue.Empty() {
				time.Sleep(time.Second)
				continue
			}
		}
		dir := getDirection(pairFormatted, interval)
		logger.Println("dir", dir)
		if dir == Up {
			rate, err := ex.GetRate(exCtx, pair)
			if err != nil {
				continue
			}
			logger.Println("current", rate)
			eps := rate * 0.0005
			buyOrder := exchange.Order{
				PushedTime: time.Now(),
				OrderTime:  time.Now().Add(30 * time.Second),
				Pair:       pair,
				Type:       "LIMIT", // todo get from exchange
				Side:       "BUY",
				Price:      round(rate+eps, info.PricePrecision),
				Amount:     round(fixedAmount/(rate+eps), info.BasePrecision),
			}
			log.Printf("order pushed %+v", buyOrder)
			queue.Push(buyOrder)

			eps = rate * 0.003
			order := exchange.Order{
				PushedTime: time.Now(),
				OrderTime:  time.Now().Add(sleepDuration * 60), // todo OrderTime???
				Pair:       pair,
				Type:       "LIMIT", // todo get from exchange
				Side:       "SELL",
				Price:      round(rate+eps, info.PricePrecision),
				Amount:     round(buyOrder.Amount, info.QuotePrecision),
			}
			log.Printf("order pushed %+v", order)
			queue.Push(order)
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
