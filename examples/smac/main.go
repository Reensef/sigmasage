package main

import (
	"context"
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
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

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

	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(5)

	for l := 1; l < 50; l++ {
		calc := func(l int) error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			startBalance := 10000.0

			deals, balanceHistory, err := tradingBotService.BacktestSMAC(
				domain.SMAInfo{
					MarketData: domain.MarketData{
						ID:           "e6123145-9665-43e0-8413-cd61b8aa9b13",
						Interval:     domain.MarketDataInterval_ONE_HOUR,
						ProviderType: domain.MarketDataProviderType_TINKOFF,
					},
					Length: l,
				},
				startBalance,
				exchange.NewMockExchange(0.0005),
				time.Now().UTC().Add(-time.Hour*24*30*24),
				time.Now().UTC(),
			)
			if err != nil {
				log.Panic(err)
			}

			if len(deals) < 2 {
				logger.Printf("Too short history for L: %d", l)
				return nil
			}

			balance := balanceHistory[len(balanceHistory)-1]
			if deals[len(deals)-1].Deal.Direction == domain.DealDirection_BUY {
				balance = balanceHistory[len(balanceHistory)-2]
			}

			logger.Printf("L: %-5d DEALS: %-5d BALANCE: %-10.2f PROFIT: %-6.2f",
				l, len(deals), balance, (balance-startBalance)/startBalance)

			return nil
		}

		g.Go(func() error {
			return calc(l)
		})
	}

	if err := g.Wait(); err != nil {
		logger.Printf("Error occurred: %v\n", err)
	} else {
		logger.Println("All workers have finished successfully.")
	}
}
