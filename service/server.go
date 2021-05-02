package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// commands
const (
	CmdRun       string = "run"
	CmdStop      string = "stop"
	CmdDelete    string = "delete"
	CmdGetState  string = "get_state"
	CmdGetLogs   string = "get_logs"
	CmdGetProfit string = "get_profit"
)

type AgentServiceHandler struct {
	srv *AgentsService
	// to stop all agents
	ctx context.Context
}

func (h *AgentServiceHandler) ServeHTTP(r *http.Request, w http.ResponseWriter) {
	switch r.Method {
	case "POST":
		id := r.Header.Get("id")
		apikey := r.Header.Get("apikey")
		apisecret := r.Header.Get("apisecret")
		baseCurr := r.Header.Get("baseCurrency")
		quoteCurr := r.Header.Get("quoteCurrency")
		interval := r.Header.Get("interval")

		err := validate(id, apikey, apisecret, baseCurr, quoteCurr, interval)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}

		err = h.srv.Create(
			id,
			apikey,
			apisecret,
			baseCurr,
			quoteCurr,
			interval,
		)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
	case "GET": // signals
		id := r.Header.Get("id")

		agent, err := h.srv.GetAgent(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}

		command := r.Header.Get("command")
		switch command {
		case CmdRun:
			go agent.Run(h.ctx)
		case CmdStop:
			h.srv.Send(Message{id, CmdStop})
		case CmdDelete:
			h.srv.Send(Message{id, CmdDelete})
			h.srv.Delete(id)
		case CmdGetState:
			h.srv.Send(Message{id, CmdGetState})
			for {
				msg := h.srv.Receive(id)
				data, err := json.Marshal(msg)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(data)
				time.Sleep(time.Second * 3)
			}
		}
	}
}

func validate(id, apikey, apisecret, baseCurr, quoteCurr, interval string) error {
	if id == "" {
		return fmt.Errorf("id empty")
	}
	if apikey == "" {
		return fmt.Errorf("apikey empty")
	}
	if apisecret == "" {
		return fmt.Errorf("apisecret empty")
	}
	if baseCurr == "" {
		return fmt.Errorf("baseCurr empty")
	}
	if quoteCurr == "" {
		return fmt.Errorf("quoteCurr empty")
	}
	if interval == "" {
		return fmt.Errorf("interval empty")
	}
	if interval != "1m" && interval != "3m" {
		return fmt.Errorf("interval not 1m, 3m")
	}
	return nil
}
