package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"time"
)

type Direction int

const (
	Neutral Direction = 0
	Up      Direction = 1
	Down    Direction = -1
)

const BatchSize = 5

var (
	app    = "Rscript"
	script = fmt.Sprintf("%s/scripts/spd.R", os.Getenv("APPDIR"))
)

func (d Direction) String() string {
	switch d {
	case 0:
		return "0"
	case 1:
		return "1"
	case -1:
		return "-1"
	}
	return "unknown"
}

type Signal struct {
	Dir  Direction
	Time time.Time
}

func getDirection(pair string, interval string) Direction {
	// TODO
	cmd := exec.Command(app, "--vanilla", script, interval, pair)

	output, err := cmd.Output()
	dir := string(output)
	log.Println(dir)
	if err != nil {
		log.Println("os exec output", err)
		return Neutral
	}
	switch dir {
	case "-1":
		return Down
	case "1":
		return Up
	case "0":
		return Neutral
	}
	return Neutral
}

type Signals []Signal

func (b *Bot) Signaller(
	ctx context.Context,
	pair, interval string,
	signalChannel chan Signal,
) {
	ex := b.exchange
	logger := log.New(os.Stdout, "[signaller] ", log.Default().Flags())

	var sleepDuration time.Duration
	switch interval {
	case "1m":
		sleepDuration = time.Minute
	case "3m":
		sleepDuration = time.Minute * 3
	case "15m":
		sleepDuration = time.Minute * 15
	}

	pairFormatted := ex.PairFormat(context.Background(), pair)
	var signal Signal
	signals := make(Signals, 0, BatchSize)
	logger.Println("started")
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		dir := getDirection(pairFormatted, interval)
		signal = Signal{dir, time.Now()}

		signals = append(signals, signal)
		signals.Flush()

		select {
		case signalChannel <- signal:
		default:

		}
		time.Sleep(sleepDuration)
	}
}

// append to csv
func (s *Signals) Flush() {
	if len(*s) < BatchSize {
		return
	}
	file, err := os.OpenFile(os.Getenv("APPDIR")+"/signals.csv", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Println("[stats]", err)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		log.Println("[stats]", err)
		return
	}
	if info.Size() == 0 {
		file.Write([]byte("time,dir,\r\n"))
	} else {
		_, err = file.Seek(info.Size(), 0)
		if err != nil {
			log.Println("[stats]", err)
			return
		}
	}

	var t, dir string
	for _, signal := range *s {
		t = signal.Time.Format(time.UnixDate)
		dir = signal.Dir.String()
		file.Write([]byte(fmt.Sprintf("%s,%s\r\n", t, dir)))
	}
	*s = (*s)[:0]
}

func round(f float64, n int) float64 {
	base := math.Pow10(n)
	return math.Round(f*base) / base
}

// 0.0003099 -> 0.000309
func roundDown(f float64, n int) float64 {
	base := math.Pow10(n)
	return math.Trunc(f*base) / base
}

/*
	or
*/
