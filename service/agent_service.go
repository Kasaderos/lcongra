package service

import (
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
	agents map[string]Agent
}

func NewAgentService(obs *Observer, rp *Reporter, lg *log.Logger) *AgentsService {
	return &AgentsService{
		MQ:       NewMQ(),
		observer: obs,
		reporter: rp,
		logger:   lg,
		agents:   make(map[string]Agent),
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
	bot := NewBot(queue, exchange, logger)

	s.Add(id)

	agent := Agent{
		MQ:            s.MQ,
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

func (s *AgentsService) GetAgent(id string) (Agent, error) {
	s.Lock()
	defer s.Unlock()
	agent, ok := s.agents[id]
	if !ok {
		return Agent{}, fmt.Errorf("no agent with id %s", id)
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
