package exchange

import "log"

type MockExchange struct {
	tocken string
}

func NewMockExchange(tocken string) *MockExchange {
	return &MockExchange{tocken: tocken}
}

func (e *MockExchange) Buy(orderInfo OrderInfo) (turnover float64, err error) {
	log.Println("Tinkoff exchange buy")
	return orderInfo.Price * float64(orderInfo.Quantity), nil
}

func (e *MockExchange) Sell(orderInfo OrderInfo) (turnover float64, err error) {
	log.Println("Tinkoff exchange sell")
	return orderInfo.Price * float64(orderInfo.Quantity), nil
}
