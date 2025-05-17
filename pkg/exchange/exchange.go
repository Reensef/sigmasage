package exchange

import "time"

type OrderRequest struct {
	InstrumentID string
	Count        int
	Price        float64
	Time         time.Time
}

type OrderResult struct {
	Count      int
	LotPrice   float64
	Commission float64
	Time       time.Time
}

type Exchanger interface {
	Buy(orderRequest OrderRequest) (orderResult OrderResult, err error)
	Sell(orderRequest OrderRequest) (orderResult OrderResult, err error)
}
