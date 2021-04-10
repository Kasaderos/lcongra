package binance

import "encoding/json"

type OrderbookResponse struct {
	LastUpdateID int64            `json:"lastUpdateId"`
	Bids         [][2]json.Number `json:"bids,string"`
	Asks         [][2]json.Number `json:"asks,string"`
}

type BalanceResponse struct {
	MakerCommission  int    `json:"makerCommission"`
	TakerCommission  int    `json:"takerCommission"`
	BuyerCommission  int    `json:"buyerCommission"`
	SellerCommission int    `json:"sellerCommission"`
	CanTrade         bool   `json:"canTrade"`
	CanWithdraw      bool   `json:"canWithdraw"`
	CanDeposit       bool   `json:"canDeposit"`
	UpdateTime       int    `json:"updateTime"`
	AccountType      string `json:"accountType"`
	Balances         []struct {
		Asset  string  `json:"asset"`
		Free   float64 `json:"free,string"`
		Locked float64 `json:"locked,string"`
	} `json:"balances"`
	Permissions []string `json:"permissions"`
}

type NewOrderResponse struct {
	Symbol              string `json:"symbol"`
	OrderID             int    `json:"orderId"`
	OrderListID         int    `json:"orderListId"`
	ClientOrderID       string `json:"clientOrderId"`
	TransactTime        int64  `json:"transactTime"`
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	TimeInForce         string `json:"timeInForce"`
	Type                string `json:"type"`
	Side                string `json:"side"`
	Fills               []struct {
		Price           string `json:"price"`
		Qty             string `json:"qty"`
		Commission      string `json:"commission"`
		CommissionAsset string `json:"commissionAsset"`
	} `json:"fills"`
}

type OpenedOrdersResponse []struct {
	Symbol              string `json:"symbol"`
	OrderID             int    `json:"orderId"`
	OrderListID         int    `json:"orderListId"`
	ClientOrderID       string `json:"clientOrderId"`
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	TimeInForce         string `json:"timeInForce"`
	Type                string `json:"type"`
	Side                string `json:"side"`
	StopPrice           string `json:"stopPrice"`
	IcebergQty          string `json:"icebergQty"`
	Time                int64  `json:"time"`
	UpdateTime          int64  `json:"updateTime"`
	IsWorking           bool   `json:"isWorking"`
	OrigQuoteOrderQty   string `json:"origQuoteOrderQty"`
}

type ExchangeInfoResponse struct {
	Timezone   string `json:"timezone"`
	ServerTime int64  `json:"serverTime"`
	RateLimits []struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		IntervalNum   int    `json:"intervalNum"`
		Limit         int    `json:"limit"`
	} `json:"rateLimits"`
	ExchangeFilters []interface{} `json:"exchangeFilters"`
	Symbols         []struct {
		Symbol                     string   `json:"symbol"`
		Status                     string   `json:"status"`
		BaseAsset                  string   `json:"baseAsset"`
		BaseAssetPrecision         int      `json:"baseAssetPrecision"`
		QuoteAsset                 string   `json:"quoteAsset"`
		QuotePrecision             int      `json:"quotePrecision"`
		QuoteAssetPrecision        int      `json:"quoteAssetPrecision"`
		BaseCommissionPrecision    int      `json:"baseCommissionPrecision"`
		QuoteCommissionPrecision   int      `json:"quoteCommissionPrecision"`
		OrderTypes                 []string `json:"orderTypes"`
		IcebergAllowed             bool     `json:"icebergAllowed"`
		OcoAllowed                 bool     `json:"ocoAllowed"`
		QuoteOrderQtyMarketAllowed bool     `json:"quoteOrderQtyMarketAllowed"`
		IsSpotTradingAllowed       bool     `json:"isSpotTradingAllowed"`
		IsMarginTradingAllowed     bool     `json:"isMarginTradingAllowed"`
		Filters                    []struct {
			FilterType       string `json:"filterType"`
			MinPrice         string `json:"minPrice,omitempty"`
			MaxPrice         string `json:"maxPrice,omitempty"`
			TickSize         string `json:"tickSize,omitempty"`
			MultiplierUp     string `json:"multiplierUp,omitempty"`
			MultiplierDown   string `json:"multiplierDown,omitempty"`
			AvgPriceMins     int    `json:"avgPriceMins,omitempty"`
			MinQty           string `json:"minQty,omitempty"`
			MaxQty           string `json:"maxQty,omitempty"`
			StepSize         string `json:"stepSize,omitempty"`
			MinNotional      string `json:"minNotional,omitempty"`
			ApplyToMarket    bool   `json:"applyToMarket,omitempty"`
			Limit            int    `json:"limit,omitempty"`
			MaxNumOrders     int    `json:"maxNumOrders,omitempty"`
			MaxNumAlgoOrders int    `json:"maxNumAlgoOrders,omitempty"`
		} `json:"filters"`
		Permissions []string `json:"permissions"`
	} `json:"symbols"`
}

type CancelOrderResponse struct {
	Symbol              string `json:"symbol"`
	OrigClientOrderID   string `json:"origClientOrderId"`
	OrderID             int    `json:"orderId"`
	OrderListID         int    `json:"orderListId"`
	ClientOrderID       string `json:"clientOrderId"`
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	TimeInForce         string `json:"timeInForce"`
	Type                string `json:"type"`
	Side                string `json:"side"`
}

type AccountResponse struct {
	MakerCommission  int    `json:"makerCommission"`
	TakerCommission  int    `json:"takerCommission"`
	BuyerCommission  int    `json:"buyerCommission"`
	SellerCommission int    `json:"sellerCommission"`
	CanTrade         bool   `json:"canTrade"`
	CanWithdraw      bool   `json:"canWithdraw"`
	CanDeposit       bool   `json:"canDeposit"`
	UpdateTime       int    `json:"updateTime"`
	AccountType      string `json:"accountType"`
	Balances         []struct {
		Asset  string `json:"asset"`
		Free   string `json:"free"`
		Locked string `json:"locked"`
	} `json:"balances"`
	Permissions []string `json:"permissions"`
}

type Klines [][]interface{}
