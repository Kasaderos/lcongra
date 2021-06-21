package binance

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/kasaderos/lcongra/exchange"
	"github.com/kasaderos/lcongra/utils/hmac"
	"github.com/kasaderos/lcongra/utils/httpf"
	"github.com/kasaderos/lcongra/utils/json"
)

// binance
var (
	Endpoint        = "https://api.binance.com"
	ApiPing         = fmt.Sprintf("%s/api/v3/ping", Endpoint)
	ApiOrderbook    = fmt.Sprintf("%s/api/v3/depth", Endpoint)
	ApiBuyCurrency  = fmt.Sprintf("%s/api/", Endpoint)
	ApiBalance      = fmt.Sprintf("%s/api/v3/account", Endpoint)
	ApiNewOrder     = fmt.Sprintf("%s/api/v3/order", Endpoint)
	ApiOpenOrders   = fmt.Sprintf("%s/api/v3/openOrders", Endpoint)
	ApiCancelOrder  = fmt.Sprintf("%s/api/v3/order", Endpoint)
	ApiKlines       = fmt.Sprintf("%s/api/v3/klines", Endpoint)
	ApiExchangeInfo = fmt.Sprintf("%s/api/v3/exchangeInfo", Endpoint)
)

const (
	MaxRequestsPerMin = 1200
)

type ExchangeMutex struct {
	mu        sync.Mutex
	lastReset time.Time
	count     int
}

func (mx *ExchangeMutex) Lock() {
	mx.mu.Lock()
	mx.count++
	if mx.count >= MaxRequestsPerMin {
		log.Println("[binance] max requests reached")
	}
}

func (mx *ExchangeMutex) Unlock() {
	now := time.Now()
	if now.Sub(mx.lastReset) > time.Minute {
		log.Println("[binance] requests per minute", mx.count)
		mx.count = 0
		mx.lastReset = now
	}
	mx.mu.Unlock()
}

type binance struct {
	Name   string
	logger *log.Logger
	mx     ExchangeMutex
}

func NewExchange(logger *log.Logger) exchange.Exchanger {
	return &binance{
		Name:   "binance",
		logger: logger,
		mx: ExchangeMutex{
			lastReset: time.Now(),
		},
	}
}

func (ex *binance) PairFormat(ctx context.Context, pair string) string {
	b, q := exchange.Currencies(pair)
	if b == "" || q == "" {
		return pair
	}
	return b + q
}

func (ex *binance) Ping(ctx context.Context) error {
	body, err := httpf.Get(ApiPing, nil, nil)
	if err != nil {
		return err
	}
	if string(body) != "{}" {
		return nil
	}
	ex.logger.Println(ex.Name, "success")
	return nil
}

func (ex *binance) GetRate(ctx context.Context, pair string) (rate float64, err error) {
	ex.mx.Lock()
	defer ex.mx.Unlock()

	pair = ex.PairFormat(ctx, pair)

	query := url.Values{}
	query.Add("symbol", pair)
	query.Add("limit", "100")
	resp, err := httpf.Get(ApiOrderbook, query, nil)
	if err != nil {
		return -1, err
	}
	// unmarshal
	var book OrderbookResponse
	err = json.Unmarshal(resp, &book)
	if err != nil {
		return -1, json.ErrJSONUnmarshal
	}
	// check length
	if len(book.Asks) == 0 {
		return 0, fmt.Errorf("asks length is zero")
	}
	if len(book.Bids) == 0 {
		return 0, fmt.Errorf("bids length is zero")
	}

	// create OrderBook
	asks := make([]exchange.Order, 0, len(book.Asks))
	for _, b := range book.Asks {
		price, err := b[0].Float64()
		if err != nil {
			return 0, json.ErrNumConversion
		}
		amount, err := b[1].Float64()
		if err != nil {
			return 0, json.ErrNumConversion
		}
		asks = append(asks, exchange.Order{
			Price:  price,
			Amount: amount,
		})
	}
	// create OrderBook
	bids := make([]exchange.Order, 0, len(book.Asks))
	for _, b := range book.Bids {
		price, err := b[0].Float64()
		if err != nil {
			return 0, json.ErrNumConversion
		}
		amount, err := b[1].Float64()
		if err != nil {
			return 0, json.ErrNumConversion
		}
		bids = append(bids, exchange.Order{
			Price:  price,
			Amount: amount,
		})
	}

	rate = exchange.NewOrderbook(asks, bids).CalcRate()
	return rate, nil
}

func (ex *binance) CreateOrder(ctx context.Context, order *exchange.Order) (id string, err error) {
	ex.mx.Lock()
	defer ex.mx.Unlock()

	pair := ex.PairFormat(ctx, order.Pair)
	keys, ok := ctx.Value("keys").(map[string]string)
	if !ok {
		return "", exchange.ErrKeysNotFound
	}

	timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
	query := url.Values{}
	query.Add("symbol", pair)
	query.Add("side", order.Side)
	query.Add("type", order.Type)
	query.Add("timestamp", timestamp)
	query.Add("price", strconv.FormatFloat(order.Price, 'f', -1, 64))
	if order.Side == "SELL" {
		query.Add("stopPrice", strconv.FormatFloat(order.StopPrice, 'f', -1, 64))
	}
	query.Add("quantity", strconv.FormatFloat(order.Amount, 'f', -1, 64))

	signature := string(hmac.SHA256(
		[]byte(query.Encode()),
		[]byte(keys["apisecret"]),
	))

	query.Set("signature", signature)

	header := http.Header{
		"X-MBX-APIKEY": []string{keys["apikey"]},
	}

	resp, err := httpf.Post(ApiNewOrder, header, query, "")
	if err != nil {
		return "", err
	}
	var newOrder NewOrderResponse
	err = json.Unmarshal(resp, &newOrder)
	if err != nil {
		return "", json.ErrJSONUnmarshal
	}

	if newOrder.ClientOrderID == "" {
		return "", exchange.ErrOrderNotCreated
	}
	// save db string(resp)
	orderID := newOrder.ClientOrderID
	return orderID, nil
}

func (ex *binance) GetBalance(ctx context.Context, currency string) (amount float64, err error) {
	ex.mx.Lock()
	defer ex.mx.Unlock()

	//pair = ex.PairFormat(ctx, pair)
	keys, ok := ctx.Value("keys").(map[string]string)
	if !ok {
		return 0.0, exchange.ErrKeysNotFound
	}
	// time now in milliseconds
	timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
	query := url.Values{}
	query.Add("recvWindow", "5000")
	query.Add("timestamp", timestamp)

	signature := string(hmac.SHA256(
		[]byte(query.Encode()),
		[]byte(keys["apisecret"])),
	)

	query.Set("signature", signature)

	header := http.Header{}
	header.Add("X-MBX-APIKEY", keys["apikey"])
	body, err := httpf.Get(ApiBalance, query, header)
	if err != nil {
		return -1, err
	}

	var balance BalanceResponse
	err = json.Unmarshal(body, &balance)
	if err != nil {
		return -1, json.ErrJSONUnmarshal
	}
	for _, b := range balance.Balances {
		if b.Locked >= 1e-8 {
			return -1, exchange.ErrBalanceLocked
		}
		if b.Asset == currency {
			ex.logger.Println(b.Asset, b.Free)
			return b.Free, nil
		}
	}
	return -1, exchange.ErrNotFound
}

func (ex *binance) OpenedOrders(ctx context.Context, pair string) (orders []exchange.Order, err error) {
	ex.mx.Lock()
	defer ex.mx.Unlock()

	pair = ex.PairFormat(ctx, pair)
	keys, ok := ctx.Value("keys").(map[string]string)
	if !ok {
		return nil, exchange.ErrKeysNotFound
	}

	timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
	query := url.Values{}
	query.Add("recvWindow", "5000")
	query.Add("timestamp", timestamp)
	query.Add("symbol", pair)

	signature := string(hmac.SHA256(
		[]byte(query.Encode()),
		[]byte(keys["apisecret"])),
	)

	query.Set("signature", signature)

	header := http.Header{}
	header.Add("X-MBX-APIKEY", keys["apikey"])
	body, err := httpf.Get(ApiOpenOrders, query, header)
	if err != nil {
		return nil, err
	}
	var ordersResp OpenedOrdersResponse
	err = json.Unmarshal(body, &ordersResp)
	if err != nil {
		return nil, err
	}
	// TODO validate
	orders = make([]exchange.Order, 0, len(ordersResp))
	for _, d := range ordersResp {
		orders = append(orders, exchange.Order{
			ID: d.ClientOrderID,
		})
	}

	return orders, nil
}

func (ex *binance) CancelOrder(ctx context.Context, pair string, orderID string) (err error) {
	ex.mx.Lock()
	defer ex.mx.Unlock()

	pair = ex.PairFormat(ctx, pair)
	keys, ok := ctx.Value("keys").(map[string]string)
	if !ok {
		return exchange.ErrKeysNotFound
	}

	timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
	query := url.Values{}
	query.Add("recvWindow", "5000")
	query.Add("timestamp", timestamp)
	query.Add("origClientOrderId", orderID)
	query.Add("symbol", pair)

	signature := string(hmac.SHA256(
		[]byte(query.Encode()),
		[]byte(keys["apisecret"])),
	)

	query.Set("signature", signature)

	header := http.Header{}
	header.Add("X-MBX-APIKEY", keys["apikey"])
	body, err := httpf.Delete(ApiCancelOrder, header, query, "")
	if err != nil {
		return err
	}
	var resp CancelOrderResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}
	if resp.Status != "CANCELED" {
		return errors.New("status not cancelled")
	}
	if resp.OrigClientOrderID != orderID {
		msg := fmt.Sprintf("order id not matched, expected %s, actual %s\n", orderID, resp.ClientOrderID)
		return errors.New(msg)
	}
	ex.logger.Println("status", "CANCELED")
	return nil
}

func (ex *binance) GetInformation(ctx context.Context, pair string) (info *exchange.Information, err error) {
	ex.mx.Lock()
	pairFormated := ex.PairFormat(ctx, pair)
	body, err := httpf.Get(ApiExchangeInfo, nil, nil)
	if err != nil {
		return nil, err
	}
	ex.mx.Unlock()

	var exchInfo ExchangeInfoResponse
	err = json.Unmarshal(body, &exchInfo)
	if err != nil {
		return nil, err
	}
	// TODO validate
	info = new(exchange.Information)
	info.Pair = pair
	exist := false
	for _, s := range exchInfo.Symbols {
		if s.Symbol == pairFormated {
			for _, f := range s.Filters {
				if f.FilterType == "PRICE_FILTER" {
					val, err := strconv.ParseFloat(f.TickSize, 64)
					if err != nil {
						return nil, err
					}
					info.PricePrecision = exchange.GetPrecision(val)
				}
				if f.FilterType == "LOT_SIZE" {
					val, err := strconv.ParseFloat(f.StepSize, 64)
					if err != nil {
						return nil, err
					}
					info.BasePrecision = exchange.GetPrecision(val)
				}
			}
			exist = true
			break
		}
	}
	if !exist {
		return nil, fmt.Errorf("pair not exist %s\n", pair)
	}

	_, quoated := exchange.Currencies(pair)
	amount, err := ex.GetBalance(ctx, quoated)
	if err != nil {
		return nil, err
	}
	info.FreeQuotedCurrency = amount
	info.CanTrade = true
	info.MinSum = 11
	ex.logger.Println("currency", quoated, "free", amount)
	return info, nil
}

// func (ex *binance) Candlesticks(ctx context.Context, pair, interval string, lastNum int) ([]exch.Candlestick, error) {
// 	query := url.Values{}
// 	query.Add("symbol", pair)
// 	query.Add("interval", interval)
// 	query.Add("limit", strconv.Itoa(lastNum))
// 	data, err := httpf.Get(ApiKlines, query, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var resp Klines
// 	err = json.Unmarshal(data, &resp)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var res []exchange.Candlestick
// 	for _, v := range resp {
// 		var candle exchange.Candlestick
// 		if len(v) < 12 {
// 			return nil, errors.New("kandlesticks invalid")
// 		}
// 		timestamp := v[0].(int64)
// 		t := time.Unix(timestamp, 0)
// 		open := v[1].(string)
// 		openV, err := strconv.ParseFloat(open, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		high := v[2].(string)
// 		highV, err := strconv.ParseFloat(high, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		low := v[3].(string)
// 		lowV, err := strconv.ParseFloat(low, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		close := v[4].(string)
// 		closeV, err := strconv.ParseFloat(close, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		vol := v[5].(string)
// 		volV, err := strconv.ParseFloat(vol, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		candle.Time = t
// 		candle.O = openV
// 		candle.H = highV
// 		candle.L = lowV
// 		candle.C = closeV
// 		candle.Volume = volV
// 		res = append(res, candle)
// 	}
// 	return res, nil
// }
