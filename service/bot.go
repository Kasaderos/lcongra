package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/kasaderos/lcongra/exchange"
)

type State int

const (
	Start State = iota
	GetSignal
	CreateOrder
	CheckOrder
	WaitOrder
	Nothing
	CancelOrder
	OpenPosition
	ClosePosition
)

const (
	AttemptsNumber = 3
	MinSum         = 11
)

var states = []string{"Start", "PopOrder", "CreateOrder", "CheckOrder", "WaitOrder", "Nothing"}

func (s State) String() string {
	return states[s]
}

type Bot struct {
	ms    sync.RWMutex // TODO
	state State

	mc       sync.RWMutex // TODO
	pair     string
	cache    float64
	count    int
	interval time.Duration

	exchange   exchange.Exchanger
	logger     *log.Logger
	info       *exchange.Information
	lastSignal Signal
	exCtx      context.Context
}

func NewBot(ex exchange.Exchanger, logger *log.Logger, pair string, exCtx context.Context) *Bot {

	return &Bot{
		exchange: ex,
		logger:   logger,
		pair:     pair,
		exCtx:    exCtx,
	}
}

func (b *Bot) GetState() State {
	b.ms.RLock()
	defer b.ms.RUnlock()
	return b.state
}

func (b *Bot) SetState(s State) {
	b.ms.Lock()
	defer b.ms.Unlock()
	b.state = s
}

func (b *Bot) GetCache() (baseSum float64, quoteSum float64, all float64) {
	base, quote := exchange.Currencies(b.pair)
	baseAmount, err := b.exchange.GetBalance(b.exCtx, base)
	if err != nil {
		b.logger.Println(err)
		return 0.0, 0.0, 0.0
	}
	quoteAmount, err := b.exchange.GetBalance(b.exCtx, quote)
	if err != nil {
		b.logger.Println(err)
		return 0.0, 0.0, 0.0
	}
	price, err := b.exchange.GetRate(b.exCtx, b.pair)
	if err != nil {
		b.logger.Println(err)
		return 0.0, 0.0, 0.0
	}
	b.ms.Lock()
	defer b.ms.Unlock()
	baseSum = price * baseAmount
	b.cache = baseSum + quoteAmount
	return baseSum, quoteAmount, b.cache
}

/*
			v---------------------------------------------------------------------------------------
	Start -> GetSignal |                                                                                    |
			    -> OpenPosition -> CreateOrder -> CheckOrder |                                      |
			    					          -> ClosePosition -> CreateOrder -> CheckOrder

*/
func (b *Bot) StartSM(ctx context.Context, msgChan <-chan string, signalChannel <-chan Signal) {
	b.SetState(Start)

	var currentOrder *exchange.Order
	var err error
	b.logger.Println("SM started")
SM:
	for {
		// b.logger.Println("state", b.state)
		select {
		case <-ctx.Done():
			b.logger.Println("deleted")
			break SM
		case cmd := <-msgChan:
			if cmd == CmdStop {
				b.SetState(Nothing)
				b.logger.Println("stopped")
			}
		default:
		}

		switch b.state {
		case Start:
			info, err := b.exchange.GetInformation(b.exCtx, b.pair)
			if err != nil {
				b.logger.Println(err)
				return
			}
			b.info = info
			b.state = GetSignal

		case GetSignal:
			select {
			case b.lastSignal = <-signalChannel:

				b.SetState(OpenPosition)
			default:
			}
		case OpenPosition:
			currentOrder, err = b.createBuyOrder()
			if err != nil {
				b.logger.Println(err)
				continue
			}
			b.SetState(CreateOrder)
		case ClosePosition:
			currentOrder = b.createSellOrder(currentOrder.Price, currentOrder.Amount)
			b.SetState(CreateOrder)
		case CreateOrder:
			b.logger.Println("state", b.state)
			var id string
			_, quote, _ := b.GetCache()
			if currentOrder.Side == "BUY" && quote < MinSum {
				b.logger.Println("not enough money in balance")
				b.SetState(Nothing)
				continue
			}
			for attempts := 0; attempts < AttemptsNumber; attempts++ {
				id, err = b.exchange.CreateOrder(b.exCtx, currentOrder)
				currentOrder.ID = id
				if err == nil {
					break
				} else {
					b.logger.Println(err)
				}
			}
			if err != nil {
				b.SetState(GetSignal)
				b.logger.Println("can't create order", err)
				continue
			}

			b.logger.Printf("order created %+v\n", currentOrder)
			b.SetState(CheckOrder)

		case CheckOrder:
			b.logger.Println("state", b.state)
			var (
				orders []exchange.Order
				err    error
			)
			for attempts := 0; attempts < AttemptsNumber; attempts++ {
				orders, err = b.exchange.OpenedOrders(b.exCtx, currentOrder.Pair)
				if err == nil {
					break
				} else {
					b.logger.Println(err)
				}
			}
			if err != nil {
				b.logger.Println("can't check order")
				time.Sleep(5 * time.Second)
				continue
			}

			exist := false
			for _, order := range orders {
				if order.ID == currentOrder.ID {
					exist = true
				}
			}
			if exist {
				b.logger.Println("order exist", currentOrder.ID, currentOrder.Side)
				time.Sleep(5 * time.Second)
			} else {
				if currentOrder.Side == "BUY" {
					b.SetState(ClosePosition)
				} else {
					b.SetState(GetSignal)
				}
				b.logger.Printf("order finished %+v\n", currentOrder)
				continue
			}

			if currentOrder.Side == "BUY" && time.Since(currentOrder.OrderTime) > b.interval {
				b.logger.Println("order not completed: side=buy")
				b.SetState(CancelOrder)
			}

		case CancelOrder:
			err = b.exchange.CancelOrder(b.exCtx, currentOrder.Pair, currentOrder.ID)
			if err != nil {
				b.logger.Println(err)
				b.SetState(GetSignal)
				continue
			}
			b.logger.Println("order cancelled:", currentOrder.ID)
		case Nothing:
			time.Sleep(10 * time.Second)
		}
	}
}

func (b *Bot) createBuyOrder() (*exchange.Order, error) {
	rate, err := b.exchange.GetRate(b.exCtx, b.pair)
	if err != nil {
		return nil, err
	}
	eps := rate * 0.0005
	buyOrder := &exchange.Order{
		PushedTime: time.Now(),
		OrderTime:  time.Now().Add(30 * time.Second),
		Pair:       b.pair,
		Type:       "LIMIT", // todo get from exchange
		Side:       "BUY",
		Price:      round(rate+eps, b.info.PricePrecision),
		Amount:     round(MinSum/(rate+eps), b.info.BasePrecision),
	}
	return buyOrder, nil
}

func (b *Bot) createSellOrder(boughtRate float64, boughtAmount float64) *exchange.Order {
	eps := boughtRate * 0.003
	order := &exchange.Order{
		PushedTime: time.Now(),
		OrderTime:  time.Now().Add(b.interval * 60), // todo OrderTime???
		Pair:       b.pair,
		Type:       "LIMIT", // todo get from exchange
		Side:       "SELL",
		Price:      round(boughtRate+eps, b.info.PricePrecision),
		Amount:     round(boughtAmount, b.info.QuotePrecision),
	}
	return order
}
