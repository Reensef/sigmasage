package main

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/Reensef/sigmasage/internal/service"
	"github.com/Reensef/sigmasage/pkg/domain"
	"github.com/Reensef/sigmasage/pkg/env"
	"github.com/Reensef/sigmasage/pkg/exchange"
	"github.com/Reensef/sigmasage/pkg/marketdata"
	"github.com/Reensef/sigmasage/pkg/techanalysis"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logDir := "bin/logs/"

	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	file, err := os.OpenFile(logDir+timestamp+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	multiWriter := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWriter, "APP: ", log.LstdFlags)

	tinkoffProvider, err := marketdata.NewTinkoffMarketDataProvider(
		env.MustString("TINKOFF_MARKET_DATA_API_TOKEN"),
	)
	if err != nil {
		log.Panic(err)
	}

	mdService, err := service.NewMarketDataService(tinkoffProvider)
	if err != nil {
		log.Panic(err)
	}

	smaProvider := techanalysis.NewSMAProvider()

	techAnalysisService := service.NewTechAnalysisService(mdService, smaProvider)
	strategyService := service.NewStrategyService(mdService, techAnalysisService)
	tradingBotService := service.NewTradingBotService(strategyService)

	intervals := []domain.MarketDataInterval{
		// domain.MarketDataInterval_ONE_MINUTE,
		// domain.MarketDataInterval_THERTY_MIN,
		domain.MarketDataInterval_ONE_HOUR,
		domain.MarketDataInterval_ONE_DAY,
		domain.MarketDataInterval_WEEK,
	}

	marketdataID := "e6123145-9665-43e0-8413-cd61b8aa9b13"
	// e6123145-9665-43e0-8413-cd61b8aa9b13 Сбер
	// 7de75794-a27f-4d81-a39b-492345813822 Яндекс
	// 02cfdf61-6298-4c0f-a9ca-9cabc82afaf3 Лукойл

	for _, interval := range intervals {
		logger.Printf("Interval: %s", marketdata.ConvertMarketDataIntervalToTime(interval))

		for l := 5; l <= 200; l += 5 {
			startBalance := 10000.0

			deals, balanceHistory, err := tradingBotService.BacktestGoldenCross(
				domain.SMAInfo{
					MarketData: domain.MarketData{
						ID:           marketdataID,
						Interval:     interval,
						ProviderType: domain.MarketDataProviderType_TINKOFF,
					},
					Length: l,
				},
				startBalance,
				exchange.NewMockExchange(0.0005, 0.0005),
				time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			)
			if err != nil {
				log.Panic(err)
			}

			if len(deals) < 2 {
				logger.Printf("Too short history for L: %d", l)
				continue
			}

			balance := balanceHistory[len(balanceHistory)-1]
			if deals[len(deals)-1].Deal.Direction == domain.DealDirection_BUY {
				balance = balanceHistory[len(balanceHistory)-2]
			}

			logger.Printf("L: %-5d DEALS: %-5d BALANCE: %-10.2f PROFIT: %-6.2f",
				l, len(deals), balance, (balance-startBalance)/startBalance)
		}
	}
}
