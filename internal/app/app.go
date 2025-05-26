package app

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/Reensef/sigmasage/internal/service"
	"github.com/Reensef/sigmasage/pkg/domain"
	"github.com/Reensef/sigmasage/pkg/env"
	"github.com/Reensef/sigmasage/pkg/marketdata"
	"github.com/Reensef/sigmasage/pkg/techanalysis"

	"github.com/joho/godotenv"
)

func Run() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logDir := "bin/logs/"

	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	file, err := os.OpenFile(
		logDir+timestamp+".log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	multiWriter := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWriter, "APP: ", log.LstdFlags)

	// tgbot, err := tgbotapi.NewBotAPI(env.MustString("TELEGRAM_API_TOKEN"))
	// if err != nil {
	// 	log.Panic(err)
	// }

	// sender := tgsender.NewSender(tgbot)

	// receiver := tgserver.NewServer(tgbot, sender)
	// receiver.Run()

	tinkoffProvider, err := marketdata.NewTinkoffMarketDataProvider(
		env.MustString("TINKOFF_MARKET_DATA_API_TOKEN"),
	)
	if err != nil {
		logger.Panic(err)
	}

	mdService, err := service.NewMarketDataService(tinkoffProvider)
	if err != nil {
		logger.Panic(err)
	}

	// candles, err := mdService.GetCandlesByTime(domain.MarketData{
	// 	ID:           "e6123145-9665-43e0-8413-cd61b8aa9b13",
	// 	Interval:     domain.MarketDataInterval_ONE_HOUR,
	// 	ProviderType: domain.MarketDataProviderType_TINKOFF,
	// },
	// 	time.Now().UTC().Add(-time.Hour*24*3),
	// 	time.Now().UTC(),
	// )
	// if err != nil {
	// 	logger.Panic(err)
	// }

	// for _, candle := range candles {
	// 	logger.Println(candle.CloseTime, candle.Close)
	// }

	// logger.Println("--------------------------------")

	smaProvider := techanalysis.NewSMAProvider()

	techAnalysisService := service.NewTechAnalysisService(mdService, smaProvider)

	strategyService := service.NewStrategyService(mdService, techAnalysisService)

	// smaHistory, err := techAnalysisService.SMAHistory(
	// 	domain.SMAInfo{
	// 		MarketData: domain.MarketData{
	// 			ID:           "e6123145-9665-43e0-8413-cd61b8aa9b13",
	// 			Interval:     domain.MarketDataInterval_ONE_HOUR,
	// 			ProviderType: domain.MarketDataProviderType_TINKOFF,
	// 		},
	// 		Length: 10,
	// 	},
	// 	time.Now().UTC().Add(-time.Hour*24*3),
	// 	time.Now().UTC(),
	// )
	// if err != nil {
	// 	log.Panic(err)
	// }

	// for _, sma := range smaHistory {
	// 	logger.Println(sma.Time, sma.Value)
	// }

	signals, err := strategyService.BacktestGoldenCross(
		domain.GoldenCrossStrategyInfo{
			Md: domain.MarketData{
				ID:           "e6123145-9665-43e0-8413-cd61b8aa9b13",
				Interval:     domain.MarketDataInterval_ONE_HOUR,
				ProviderType: domain.MarketDataProviderType_TINKOFF,
			},
			ShortLength: 50,
			LongLength:  100,
		},
		time.Now().UTC().Add(-time.Hour*24*28*6),
		time.Now().UTC(),
	)

	if err != nil {
		logger.Panic(err)
	}
	for _, signal := range signals {
		logger.Println(signal.Time, signal.SignalType)
	}

	// Need graceful shutdown
	// for {

	// }
}
