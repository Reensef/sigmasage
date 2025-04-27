package exchange

type OrderInfo struct {
	InstrumentID string
	Quantity     int64
	Price        float64
}

type Exchanger interface {
	Buy(orderInfo OrderInfo) (turnover float64, err error)
	Sell(orderInfo OrderInfo) (turnover float64, err error)
}
