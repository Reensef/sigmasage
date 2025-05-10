package app

import (
	"log"
	"time"

	"github.com/Reensef/sigmasage/internal/telegram/tgsender"
	"github.com/Reensef/sigmasage/internal/telegram/tgserver"
	"github.com/Reensef/sigmasage/pkg/env"
	"github.com/Reensef/sigmasage/pkg/marketdata"
	"github.com/Reensef/sigmasage/pkg/strategy"
	"github.com/Reensef/sigmasage/pkg/techanalysis"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func Run() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	tgbot, err := tgbotapi.NewBotAPI(env.MustString("TELEGRAM_API_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	sender := tgsender.NewSender(tgbot)

	receiver := tgserver.NewServer(tgbot, sender)
	receiver.Run()

	md, err := marketdata.NewMarketDataService(env.MustString("TINKOFF_MARKET_DATA_API_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	techAnalysis, err := techanalysis.NewTechAnalysisService(md)
	if err != nil {
		log.Panic(err)
	}

	strategyService, err := strategy.NewStrategyService(md, techAnalysis)
	if err != nil {
		log.Panic(err)
	}

	from := time.Now().UTC().Truncate(time.Minute).Add(-time.Hour * 24 * 30)
	to := time.Now().UTC().Truncate(time.Minute).Add(-time.Hour * 24)

	signals, err := strategyService.SMACBacktest(techanalysis.SMAInfo{
		MarketData: marketdata.MarketData{
			ID:           "e6123145-9665-43e0-8413-cd61b8aa9b13",
			Interval:     marketdata.ONE_MINUTE,
			ProviderType: marketdata.TINKOFF,
		},
		Length: 50,
	}, from, to)

	if err != nil {
		log.Panic(err)
	}

	for _, signal := range signals {
		log.Printf("Signal: %v", signal)
	}

	// result, err := techAnalysis.SMAHistory(techanalysis.SMAInfo{
	// 	MarketData: marketdata.MarketData{
	// 		ID:           "e6123145-9665-43e0-8413-cd61b8aa9b13",
	// 		Interval:     marketdata.ONE_MINUTE,
	// 		ProviderType: marketdata.TINKOFF,
	// 	},
	// 	Length: 50,
	// }, time.Now().UTC().Truncate(time.Minute).Add(-time.Hour), time.Now().UTC().Truncate(time.Minute))

	// if err != nil {
	// 	log.Panic(err)
	// }

	// for _, sma := range result {
	// 	log.Printf("SMA: %v", sma)
	// }

	// strategyService, err := strategy.NewStrategyService(md, techAnalysis)

	// Need graceful shutdown
	for {

	}

	////////////////// Test stream and get candles //////////////////
	// md, err := marketdata.NewMarketDataService("t.b8rY7Kv7HppxKdwCynTNbZbJ6kCzR09EbEQf5gAsJhGIOVU1oteZAit8eRdInD1pFxC8h8PvnAtumjuvcDn3pg")
	// if err != nil {
	// 	log.Fatalf("Error creating marketdata: %v", err)
	// }

	// ch, err := md.SubscribeCandles(marketdata.MarketData{
	// 	ID:           "e6123145-9665-43e0-8413-cd61b8aa9b13",
	// 	Interval:     marketdata.ONE_MINUTE,
	// 	ProviderType: marketdata.TINKOFF,
	// })

	// if err != nil {
	// 	log.Fatalf("Error subscribing to candles: %v", err)
	// }

	// var wg sync.WaitGroup
	// wg.Add(2)

	// go func() {
	// 	defer wg.Done()
	// 	for candle := range ch {
	// 		log.Printf("Candle: %v", candle)
	// 	}
	// }()

	// techAnalysis, err := techanalysis.NewTechAnalysisService(md)
	// if err != nil {
	// 	log.Fatalf("Error creating tech analysis: %v", err)
	// }

	// techCh, err := techAnalysis.SubscribeSMA(techanalysis.SMAInfo{
	// 	MarketData: marketdata.MarketData{
	// 		ID:           "e6123145-9665-43e0-8413-cd61b8aa9b13",
	// 		Interval:     marketdata.ONE_MINUTE,
	// 		ProviderType: marketdata.TINKOFF,
	// 	},
	// 	Length: 10,
	// })

	// if err != nil {
	// 	log.Fatalf("Error subscribing to SMA: %v", err)
	// }

	// go func() {
	// 	defer wg.Done()
	// 	for sma := range techCh {
	// 		log.Printf("SMA: %v", sma)
	// 	}
	// }()

	// wg.Wait()

	// if err != nil {
	// 	log.Fatalf("Error getting candles: %v", err)
	// }

	// for _, candle := range candles {
	// 	log.Printf("Candle: %v", candle)
	// }

	////////////////// Test bot //////////////////
	// md, err := marketdata.NewTinkoffMarketdata("t.b8rY7Kv7HppxKdwCynTNbZbJ6kCzR09EbEQf5gAsJhGIOVU1oteZAit8eRdInD1pFxC8h8PvnAtumjuvcDn3pg")
	// if err != nil {
	// 	log.Fatalf("Error creating marketdata: %v", err)
	// }

	// bot := tradingbots.NewSMACandlesBot(
	// 	md,
	// 	exchange.NewMockExchange(),
	// 	"b71bd174-c72c-41b0-a66f-5f9073e0d1f5",
	// 	marketdata.ONE_DAY,
	// 	50,
	// 	1000,
	// )

	// bot.RunHistory(time.Now().Add(-time.Hour*24*30*12), time.Now())

}
