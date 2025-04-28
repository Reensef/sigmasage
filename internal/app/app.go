package app

import (
	"log"
	"time"

	"github.com/Reensef/sigmasage/pkg/exchange"
	"github.com/Reensef/sigmasage/pkg/marketdata"
	"github.com/Reensef/sigmasage/pkg/tradingbots"
	"github.com/joho/godotenv"
)

func Run() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables from system")
	}

	////////////////// Test stream and get candles //////////////////
	// md := marketdata.NewTinkoffMarketdata("t.b8rY7Kv7HppxKdwCynTNbZbJ6kCzR09EbEQf5gAsJhGIOVU1oteZAit8eRdInD1pFxC8h8PvnAtumjuvcDn3pg")

	// ch, err := md.SubscribeCandles(marketdata.MarketDataInfo{
	// 	IntrumentID: "BBG004730N88",
	// 	Interval:    marketdata.ONE_MINUTE,
	// })

	// if err != nil {
	// 	log.Fatalf("Error subscribing to candles: %v", err)
	// }

	// for candle := range ch {
	// 	log.Printf("Candle: %v", candle)
	// }

	// candles, err := md.GetCandles(marketdata.MarketDataInfo{
	// 	IntrumentID: "BBG004730N88",
	// 	Interval:    marketdata.ONE_MINUTE,
	// }, time.Now().Add(-time.Hour*24), time.Now())

	// if err != nil {
	// 	log.Fatalf("Error getting candles: %v", err)
	// }

	// for _, candle := range candles {
	// 	log.Printf("Candle: %v", candle)
	// }

	////////////////// Test bot //////////////////
	md, err := marketdata.NewTinkoffMarketdata("t.b8rY7Kv7HppxKdwCynTNbZbJ6kCzR09EbEQf5gAsJhGIOVU1oteZAit8eRdInD1pFxC8h8PvnAtumjuvcDn3pg")
	if err != nil {
		log.Fatalf("Error creating marketdata: %v", err)
	}

	bot := tradingbots.NewSMACandlesBot(
		md,
		exchange.NewMockExchange(),
		"b71bd174-c72c-41b0-a66f-5f9073e0d1f5",
		marketdata.ONE_DAY,
		50,
		1000,
	)

	bot.RunHistory(time.Now().Add(-time.Hour*24*30*12), time.Now())

}
