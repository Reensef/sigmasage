package exchange

import "log"

type TinkoffExchange struct {
	tocken string
}

func New(tocken string) *TinkoffExchange {
	return &TinkoffExchange{tocken: tocken}
}

func (e *TinkoffExchange) Buy() {
	log.Println("Tinkoff exchange buy")
}

func (e *TinkoffExchange) Sell() {
	log.Println("Tinkoff exchange sell")
}
