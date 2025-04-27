package exchange

import "log"

type TinkoffExchange struct {
	tocken string
}

func NewTinkoffExchange(tocken string) *TinkoffExchange {
	return &TinkoffExchange{tocken: tocken}
}

func (e *TinkoffExchange) Buy(orderInfo OrderInfo) (price float64, err error) {
	log.Println("Tinkoff exchange buy")
	return 0, nil
}

func (e *TinkoffExchange) Sell(orderInfo OrderInfo) (price float64, err error) {
	log.Println("Tinkoff exchange sell")
	return 0, nil
}
