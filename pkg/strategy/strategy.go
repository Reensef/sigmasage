package strategy

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/Reensef/sigmasage/pkg/marketdata"
)

// Конфиг стратегии пересечения скользящих средних (SimpleMovingAvgCross)
type SMACStrategyConfig struct {
	MarketData marketdata.MarketData
	SMAPeriod  int
}

// Сигнал стратегии пересечения скользящих средних (SimpleMovingAvgCross)
type SMACStrategySignal struct{}

type StrategyService struct {
	mdService               *marketdata.MarketDataService
	smacStrategySubscribers map[SMACStrategyConfig][]chan SMACStrategySignal
	smacStrategyRunners     map[SMACStrategyConfig][]context.CancelFunc
}

func NewStrategyService(mdService *marketdata.MarketDataService) (*StrategyService, error) {
	return &StrategyService{
		mdService: mdService,
	}, nil
}

func (s *StrategyService) SubscribeSMACStrategy(
	config SMACStrategyConfig,
) (<-chan SMACStrategySignal, error) {
	ch := make(chan SMACStrategySignal, 100)

	if _, exists := s.smacStrategySubscribers[config]; !exists {
		s.smacStrategySubscribers[config] = make([]chan SMACStrategySignal, 0)

		err := s.runSMACStrategy(config)
		if err != nil {
			return nil, err
		}
	}
	s.smacStrategySubscribers[config] = append(s.smacStrategySubscribers[config], ch)

	return ch, nil
}

func (s *StrategyService) UnsubscribeSMACStrategy(
	config SMACStrategyConfig,
	ch <-chan SMACStrategySignal,
) error {
	ok := false

	if subscribers, exists := s.smacStrategySubscribers[config]; exists {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				s.smacStrategySubscribers[config] = slices.Delete(subscribers, i, i+1)
				close(subscriber)
				ok = true
				break
			}
		}

		if len(s.smacStrategySubscribers[config]) == 0 {
			delete(s.smacStrategySubscribers, config)
		}
	}

	if ok {
		return nil
	} else {
		return fmt.Errorf("subscribe not found")
	}
}

func (s *StrategyService) runSMACStrategy(
	config SMACStrategyConfig,
) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.smacStrategyRunners[config] = append(s.smacStrategyRunners[config], cancel)

	go func(ctx context.Context) {
		candlesChan, err := s.mdService.SubscribeCandles(config.MarketData)
		if err != nil {
			return
		}
		for {
			select {
			case <-ctx.Done():
				s.mdService.UnsubscribeCandles(config.MarketData, candlesChan)
				return
			case candle := <-candlesChan:
				log.Printf("candle: %v", candle)
				// Тут должен быть код, который будет обрабатывать свечу и индикаторы
			}
		}
	}(ctx)

	return nil
}
