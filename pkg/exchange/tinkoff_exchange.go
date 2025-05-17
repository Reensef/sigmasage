package exchange

import "log"

type TinkoffExchange struct {
	tocken string
}

func NewTinkoffExchange(tocken string) *TinkoffExchange {
	return &TinkoffExchange{tocken: tocken}
}

func (e *TinkoffExchange) Buy(orderRequest OrderRequest) (orderResult OrderResult, err error) {
	log.Println("Tinkoff exchange buy")
	return OrderResult{
		Count:      orderRequest.Count,
		LotPrice:   orderRequest.Price,
		Commission: 0,
		Time:       orderRequest.Time,
	}, nil
}

func (e *TinkoffExchange) Sell(orderRequest OrderRequest) (orderResult OrderResult, err error) {
	log.Println("Tinkoff exchange sell")
	return OrderResult{
		Count:      orderRequest.Count,
		LotPrice:   orderRequest.Price,
		Commission: 0,
		Time:       orderRequest.Time,
	}, nil
}
