package service

import (
	"net/http"
	"strings"

	"github.com/uber-go/zap"
)

type Reporter struct {
	url    string
	logger *zap.Logger
}

func NewReporter(url string, logger *zap.Logger) *Reporter {
	return &Reporter{url, logger}
}

func (r *Reporter) report(chanMsgs <-chan string) {
	for msg := range chanMsgs {
		buf := strings.NewReader(msg)
		resp, err := http.Post(r.url, "application/json", buf)
		if err != nil {
			r.logger.Error("[reporter] report failed, msg=", msg)
			continue
		}
		if resp.StatusCode != http.StatusAccepted {
			r.logger.Error("[reporter] request not accepted, msg=", msg)
		}
	}
}
