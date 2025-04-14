package exchange

import "log"

type TinkoffExchange struct {
}

func New() *TinkoffExchange {
	return &TinkoffExchange{}
}

func (e *TinkoffExchange) Buy() {
	log.Println("Tinkoff exchange buy")
}

func (e *TinkoffExchange) Sell() {
	log.Println("Tinkoff exchange sell")
}
