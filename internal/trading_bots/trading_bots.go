package tradingbots

type Exchange interface {
	Buy() // тикет и количество
	Sell() // тикет и количество
}

type MarketData interface {
	SubscribeCandle(instrumentID string) ()
}

func SMA() {
	
}