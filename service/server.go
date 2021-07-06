package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/kasaderos/lcongra/exchange"
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

func NewAgentServiceHandler(srv *AgentsService, ctx context.Context) *AgentServiceHandler {
	return &AgentServiceHandler{
		srv,
		ctx,
	}
}

type RequestParams struct {
	ID        string  `json:"id,omitempty"`
	Pair      string  `json:"pair,omitempty"`
	Interval  string  `json:"interval,omitempty"`
	State     string  `json:"state,omitempty"`
	Cache     float64 `json:"cache,omitempty"`
	Apikey    string  `json:"apikey,omitempty"`
	Apisecret string  `json:"apisecret,omitempty"`
}

func getParams(r *http.Request) (*RequestParams, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println(string(data))
	req := new(RequestParams)
	err = json.Unmarshal(data, req)
	return req, err
}

func (h *AgentServiceHandler) sendlistBots(req *RequestParams, w http.ResponseWriter) {
	info := h.srv.GetListInfo()
	data, err := json.Marshal(info)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(data)
}

func (h *AgentServiceHandler) createBot(req *RequestParams, w http.ResponseWriter) {
	h.srv.logger.Println("createBot")
	base, quote := exchange.Currencies(req.Pair)
	err := validate(req.ID, req.Apikey, req.Apisecret, base, quote, req.Interval)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	h.srv.logger.Println("validated")
	err = h.srv.Create(req.ID, req.Apikey, req.Apisecret, base, quote, req.Interval)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AgentServiceHandler) deleteBot(req *RequestParams, w http.ResponseWriter) {
	err := h.srv.stopAndDeleteAgent(req.ID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AgentServiceHandler) runBot(req *RequestParams, w http.ResponseWriter) {
	err := h.srv.RunAgent(req.ID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AgentServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			h.srv.logger.Println(err)
			return
		}
	}()
	params, err := getParams(r)
	if err != nil {
		h.srv.logger.Println("server: json:", err)
		return
	}
	// fmt.Println(r.URL.Path)
	switch r.URL.Path {
	case "/create":
		h.createBot(params, w)
	case "/list":
		h.sendlistBots(params, w)
	case "/run":
		h.runBot(params, w)
	case "/stop":
	case "/delete":
		h.deleteBot(params, w)
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
	if interval != "1m" && interval != "3m" && interval != "15m" {
		return fmt.Errorf("interval not 1m, 3m, 15m")
	}
	return nil
}
