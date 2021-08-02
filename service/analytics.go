package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"time"

	"github.com/kasaderos/lcongra/exchange"
)

type Direction int

const (
	Stay Direction = 0
	Up   Direction = 1
	Up2  Direction = 2
	Down Direction = -1
)

const BatchSize = 5

var (
	app    = "Rscript"
	script = fmt.Sprintf("%s/scripts/la1.R", os.Getenv("APPDIR"))
)

func (d Direction) String() string {
	switch d {
	case 0:
		return "0"
	case 1:
		return "1"
	case 2:
		return "2"
	case -1:
		return "-1"
	}
	return "unknown"
}

type Signal struct {
	Dir  Direction
	Time time.Time
}

// ...  \n <- index
// <dir>\n     => <dir>
func getResult(out []byte) string {
	p := 0
	if (len(out) < 2) {
		return "0"
	}
	for i := len(out) - 2; i >= 0; i-- {
		if out[i] == '\n' {
			p = i
			break
		}
	}
	return string(out[p+1 : len(out)-1])
}

func getDirection(pair string, interval string) Direction {
	// TODO
	cmd := exec.Command(app, "--vanilla", script, interval, pair)

	output, err := cmd.Output()
	if err != nil {
		log.Println("os exec output", err)
		return Stay
	}
	dir := getResult(output)
	log.Printf("[%s] %s %s\n", pair, output, dir)
	switch dir {
	case "-1":
		return Down
	case "1":
		return Up
	case "2":
		return Up2
	case "0":
		return Stay
	}
	return Stay
}

type Signals []Signal

func Autotrade(
	ctx context.Context,
	pair, interval string,
	ex exchange.Exchanger,
	signalChannel chan Signal,
) {
	logger := log.New(os.Stdout, "[autotrade] ", log.Default().Flags())

	if interval != "15m" {
		logger.Println("invalid interval")
		return
	}

	sleepDuration := time.Minute * 5

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
		if dir == Stay {
			time.Sleep(time.Second * 5)
			continue
		}
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
