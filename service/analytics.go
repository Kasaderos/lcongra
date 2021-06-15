package service

import (
	"context"
	"log"
	"math"
	"os/exec"
	"time"
	"strings"
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
	if err != nil {
		log.Println("os exec output", err)
		return Stay
	}
	res := string(output)
	log.Println(res)
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
	queue *OrderQueue,
	ex exchange.Exchanger,
) {
	var sleepDuration time.Duration
	switch interval {
	case "3m":
		sleepDuration = time.Minute * 3
	case "1m":
		sleepDuration = time.Second * 20 
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
		// log.Println("dir", dir)
		if dir == Up {
			rate, err := ex.GetRate(pair)
			if err != nil {
				continue
			}
			log.Println("current", rate)
			eps := rate * 0.0005
			buyOrder := exchange.Order{
				PushedTime: time.Now(),
				OrderTime:  time.Now().Add(10 * time.Second),
				Pair:       pair,
				Type:       "LIMIT", // todo get from exchange
				Side:       "BUY",
				Price:      round(rate+eps, 2),
				Amount:     round(fixedAmount/(rate+eps), 8),
			}
			// log.Printf("order pushed %+v", buyOrder)
			queue.Push(buyOrder)

			eps = rate * 0.003
			order := exchange.Order{
				PushedTime: time.Now(),
				OrderTime:  time.Now().Add(sleepDuration * 60), // todo OrderTime???
				Pair:       pair,
				Type:       "LIMIT", // todo get from exchange
				Side:       "SELL",
				Price:      round(rate+eps, 2),
				Amount:     round(buyOrder.Amount-1e-6, 8),
			}
			// log.Printf("order pushed %+v", order)
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
