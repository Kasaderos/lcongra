package service

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type Reporter struct {
	url      string
	logger   *log.Logger
	chanMsgs <-chan string
}

func NewReporter(url string, logger *log.Logger, chanMsgs <-chan string) *Reporter {
	return &Reporter{url, logger, chanMsgs}
}

func (r *Reporter) Report(ctx context.Context) {
	for msg := range r.chanMsgs {
		select {
		case <-ctx.Done():
			return
		default:
		}
		buf := strings.NewReader(msg)
		resp, err := http.Post(r.url, "application/json", buf)
		if err != nil {
			r.logger.Printf("report failed, msg=%s\n", msg)
			continue
		}
		if resp.StatusCode != http.StatusAccepted {
			r.logger.Printf("request not accepted, msg=%s\n", msg)
		}
	}
}
