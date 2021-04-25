package service

import (
	"context"
	"log"
	"os/exec"
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
	cmd := exec.Command(app, "--vanilla", "../../scripts/la1.R", interval, pair)

	output, err := cmd.Output()
	if err != nil {
		log.Println("os exec output", err)
		return Stay
	}
	dir := string(output)
	switch dir {
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
	queue *OrderQueue,
	ex exchange.Exchanger,
) {
	var sleepDuration time.Duration
	switch interval {
	case "3m":
		sleepDuration = time.Minute * 3
	}

	pair = ex.PairFormat(pair)
	minAmount := 11.0
	fixedAmount := minAmount
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
		dir := getDirection(pair, interval)
		if dir == Up {
			rate, err := ex.GetRate(pair)
			if err != nil {
				continue
			}
			eps := rate + rate*0.001
			order := exchange.Order{
				PushedTime: time.Now(),
				Pair:       pair,
				Type:       "LIMIT", // todo get from exchange
				Side:       "BUY",
				Price:      rate + eps,
				Amount:     fixedAmount,
			}
			queue.Push(order)

			eps = rate + rate*0.005
			order = exchange.Order{
				PushedTime: time.Now(),
				OrderTime:  time.Now().Add(sleepDuration * 60), // todo OrderTime???
				Pair:       pair,
				Type:       "LIMIT", // todo get from exchange
				Side:       "SELL",
				Price:      rate + eps,
				Amount:     fixedAmount,
			}
			queue.Push(order)
		}

		time.Sleep(sleepDuration)
	}
}

/*
	or
*/
