package service

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type Reporter struct {
	url      string
	logger   *zap.Logger
	chanMsgs <-chan string
}

func NewReporter(url string, logger *zap.Logger, chanMsgs <-chan string) *Reporter {
	return &Reporter{url, logger, chanMsgs}
}

func (r *Reporter) Report() {
	for msg := range r.chanMsgs {
		buf := strings.NewReader(msg)
		resp, err := http.Post(r.url, "application/json", buf)
		if err != nil {
			r.logger.Error("[reporter] report failed")
			continue
		}
		if resp.StatusCode != http.StatusAccepted {
			r.logger.Error("[reporter] request not accepted")
		}
	}
}
