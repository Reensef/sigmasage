package exchange

type MockExchange struct {
}

func NewMockExchange() *MockExchange {
	return &MockExchange{}
}

func (e *MockExchange) Buy(orderInfo OrderInfo) (turnover float64, err error) {
	// log.Println("Mock exchange buy", orderInfo.Price)
	return orderInfo.Price * float64(orderInfo.Count), nil
}

func (e *MockExchange) Sell(orderInfo OrderInfo) (turnover float64, err error) {
	// log.Println("Mock exchange sell", orderInfo.Price)
	return orderInfo.Price * float64(orderInfo.Count), nil
}
