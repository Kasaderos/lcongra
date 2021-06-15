package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	ex "github.com/kasaderos/lcongra/exchange/fake"
)

type AgentsService struct {
	*MQ
	observer *Observer
	reporter *Reporter
	logger   *log.Logger
	sync.Mutex
	agents map[string]*Agent
}

func NewAgentService(obs *Observer, rp *Reporter, lg *log.Logger) *AgentsService {
	mq := NewMQ()
	mq.AddQueue("master")
	return &AgentsService{
		MQ:       mq,
		observer: obs,
		reporter: rp,
		logger:   lg,
		agents:   make(map[string]*Agent),
	}
}

func (s *AgentsService) Create(
	id string,
	apikey, apisecret string,
	baseCurr, quoteCurr string,
	interval string,
) error {
	prefix := fmt.Sprintf("[exchange-%s] ", id)
	logger := log.New(os.Stdout, prefix, log.Default().Flags())
	// set keys when real exchange
	exchange := ex.NewExchange(logger)
	queue := NewOrderQueue()
	prefix = fmt.Sprintf("[%s] ", id)
	logger = log.New(os.Stdout, prefix, log.Default().Flags())
	bot := NewBot(queue, exchange, logger, baseCurr+"-"+quoteCurr)

	// s.AddQueue(id)

	agent := &Agent{
		// MQ:            s.MQ,
		ID:            id,
		bot:           bot,
		queue:         queue,
		exchange:      exchange,
		baseCurrency:  baseCurr,
		quoteCurrency: quoteCurr,
		interval:      interval,
	}
	_, err := s.GetAgent(id)
	if err == nil {
		return fmt.Errorf("bot exist with id %s", id)
	}
	s.Lock()
	defer s.Unlock()
	s.agents[id] = agent
	return nil
}

func (s *AgentsService) GetAgent(id string) (*Agent, error) {
	s.Lock()
	defer s.Unlock()
	agent, ok := s.agents[id]
	if !ok {
		return nil, fmt.Errorf("no agent with id %s", id)
	}
	return agent, nil
}

// To form new routings
func (s *AgentsService) GetIDs() []string {
	s.Lock()
	defer s.Unlock()
	ids := make([]string, 0, len(s.agents))
	for k := range s.agents {
		ids = append(ids, k)
	}
	return ids
}

func (s *AgentsService) GetListInfo() []AgentInfo {
	s.RLock()
	defer s.RUnlock()
	agents := make([]AgentInfo, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, AgentInfo{
			ID:       agent.ID,
			Pair:     agent.baseCurrency + "-" + agent.quoteCurrency,
			Interval: agent.interval,
			State:    agent.bot.GetState().String(),
			Cache:    agent.bot.GetCache(),
		})
	}
	return agents
}

func (s *AgentsService) Delete(id string) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.agents[id]
	if !ok {
		s.logger.Println("delete id:", id, "not found")
		return
	}
	delete(s.agents, id)
}

func (s *AgentsService) stopAndDeleteAgent(id string) error {
	ag, err := s.GetAgent(id)
	if err != nil {
		return err
	}
	ag.cancel()
	s.Delete(id)
	return nil
}

func (s *AgentsService) RunAgent(id string) error {
	ag, err := s.GetAgent(id)
	if err != nil {
		s.logger.Println("agent not found", id)
		return err
	}
	ag.mu.RLock()
	if ag.ctx != nil {
		_, ok := ag.ctx.Deadline()
		if !ok {
			return errors.New(id + " is running")
		}
	}
	ag.mu.RUnlock()

	msgChan := make(chan string)

	ctx, quit := context.WithCancel(context.Background())
	ag.mu.Lock()
	ag.ctx = ctx
	ag.cancel = quit
	ag.mu.Unlock()
	s.logger.Println("bot started", id)
	go func() {

		tradeCtx, cancel := context.WithCancel(context.Background())
		// for fake exchange, we need
		go ex.Update(tradeCtx, ag.exchange)

		go ag.bot.StartSM(tradeCtx, msgChan)
		go Autotrade(
			tradeCtx,
			fmt.Sprintf("%s-%s", ag.baseCurrency, ag.quoteCurrency),
			ag.interval,
			ag.bot.queue,
			ag.exchange,
		)

		for {
			select {
			case <-ctx.Done():
				cancel()
				ag.bot.logger.Println("stopped")
				return
				// default:
			}

			// msg := ag.MQ.Receive(ag.ID)
			// switch msg.Data {
			// case CmdDelete:
			// 	cancel()
			// 	return
			// default:
			// 	if msg.Data != "no message" {
			// 		msgChan <- msg.Data
			// 	}
			// }
			// time.Sleep(2 * time.Second)
		}
	}()
	return nil
}
