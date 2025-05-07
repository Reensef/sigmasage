package strategy

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Reensef/sigmasage/pkg/marketdata"
	"github.com/Reensef/sigmasage/pkg/techanalysis"
)

// Сигнал стратегии пересечения скользящих средних (SimpleMovingAvgCross)
type SMACSignal int

const (
	SMACSignal_UPWARDS_CROSS SMACSignal = iota
	SMACSignal_DOWNWARDS_CROSS
)

type StrategyService struct {
	mdService               *marketdata.MarketDataService
	techAnalysisService     *techanalysis.TechAnalysisService
	smacStrategySubscribers map[techanalysis.SMAInfo][]chan SMACSignal
	smacStrategyRunners     map[techanalysis.SMAInfo][]context.CancelFunc
}

func NewStrategyService(
	mdService *marketdata.MarketDataService,
	techAnalysisService *techanalysis.TechAnalysisService,
) (*StrategyService, error) {
	return &StrategyService{
		mdService:               mdService,
		techAnalysisService:     techAnalysisService,
		smacStrategySubscribers: make(map[techanalysis.SMAInfo][]chan SMACSignal),
		smacStrategyRunners:     make(map[techanalysis.SMAInfo][]context.CancelFunc),
	}, nil
}

func (s *StrategyService) SubscribeSMACStrategy(
	config techanalysis.SMAInfo,
) (<-chan SMACSignal, error) {
	ch := make(chan SMACSignal, 100)

	if _, exists := s.smacStrategySubscribers[config]; !exists {
		s.smacStrategySubscribers[config] = make([]chan SMACSignal, 0)

		err := s.runSMACStrategy(config)
		if err != nil {
			return nil, err
		}
	}
	s.smacStrategySubscribers[config] = append(s.smacStrategySubscribers[config], ch)

	return ch, nil
}

func (s *StrategyService) UnsubscribeSMACStrategy(
	config techanalysis.SMAInfo,
	ch <-chan SMACSignal,
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
	config techanalysis.SMAInfo,
) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.smacStrategyRunners[config] = append(s.smacStrategyRunners[config], cancel)

	candlesChan, err := s.mdService.SubscribeCandles(config.MarketData)
	if err != nil {
		return err
	}

	smaChan, err := s.techAnalysisService.SubscribeSMA(config)
	if err != nil {
		return err
	}

	go func(ctx context.Context) {
		firstCandle, firstSMA := s.syncSMACData(candlesChan, smaChan, ctx)

		if firstCandle == nil || firstSMA == nil {
			s.mdService.UnsubscribeCandles(config.MarketData, candlesChan)
			return
		}

		isCandleCloseAbove := firstCandle.Close > firstSMA.Value
		recvCandle := *firstCandle
		recvSMA := *firstSMA

		makeDecision := func(candle marketdata.Candle, sma techanalysis.SMA) {
			if candle.Close > sma.Value {
				if !isCandleCloseAbove {
					s.notifySMACSubscribers(config, SMACSignal_UPWARDS_CROSS)
					isCandleCloseAbove = true
				}
			} else if candle.Close < recvSMA.Value {
				if isCandleCloseAbove {
					s.notifySMACSubscribers(config, SMACSignal_DOWNWARDS_CROSS)
					isCandleCloseAbove = false
				}
			}
		}

		for {
			select {
			case <-ctx.Done():
				s.mdService.UnsubscribeCandles(config.MarketData, candlesChan)
				return
			case candle := <-candlesChan:
				recvCandle = candle

				if time.Until(candle.EndTime).Abs() < time.Second {
					if recvCandle.EndTime == recvSMA.Time {
						makeDecision(recvCandle, recvSMA)
					}
				}
			case sma := <-smaChan:
				recvSMA = sma

				if time.Until(recvSMA.Time).Abs() < time.Second {
					if recvCandle.EndTime == recvSMA.Time {
						makeDecision(recvCandle, recvSMA)
					}
				}
			}
		}
	}(ctx)

	return nil
}

func (s *StrategyService) syncSMACData(
	candlesChan <-chan marketdata.Candle,
	smaChan <-chan techanalysis.SMA,
	ctx context.Context,
) (*marketdata.Candle, *techanalysis.SMA) {
	var candle *marketdata.Candle
	var sma *techanalysis.SMA

	for {
		select {
		case <-ctx.Done():
			return candle, sma
		case recvCandle := <-candlesChan:
			candle = &recvCandle

			if sma != nil && candle.EndTime == sma.Time {
				return candle, sma
			}
		case recvSMA := <-smaChan:
			sma = &recvSMA
			if candle != nil && candle.EndTime == sma.Time {
				return candle, sma
			}
		}
	}
}

func (s *StrategyService) notifySMACSubscribers(info techanalysis.SMAInfo, signal SMACSignal) {
	subscribers, exists := s.smacStrategySubscribers[info]
	if !exists {
		return
	}

	for _, subscriber := range subscribers {
		subscriber <- signal
	}
}
