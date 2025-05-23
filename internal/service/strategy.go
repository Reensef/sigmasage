package service

import (
	"fmt"
	"log"
	"time"

	"github.com/Reensef/sigmasage/pkg/domain"
	"github.com/Reensef/sigmasage/pkg/strategy"
)

type StrategyService struct {
	mdService           *MarketDataService
	techAnalysisService *TechAnalysisService
	smacStrategy        *strategy.SMACStrategy
	goldenCrossStrategy *strategy.GoldenCrossStrategy
	smaSignalToSMA      map[<-chan domain.SMACSignal]<-chan domain.SMA
	smaSignalToCandles  map[<-chan domain.SMACSignal]<-chan domain.Candle
}

func NewStrategyService(mdService *MarketDataService, techAnalysisService *TechAnalysisService) *StrategyService {
	return &StrategyService{
		mdService:           mdService,
		techAnalysisService: techAnalysisService,
		smaSignalToSMA:      make(map[<-chan domain.SMACSignal]<-chan domain.SMA),
		smaSignalToCandles:  make(map[<-chan domain.SMACSignal]<-chan domain.Candle),
	}
}

func (s *StrategyService) SubscribeSMAC(
	info domain.SMAInfo,
) (<-chan domain.SMACSignal, error) {
	candleChan, err := s.mdService.SubscribeCandles(info.MarketData)
	if err != nil {
		return nil, err
	}

	smaChan, err := s.techAnalysisService.SubscribeSMA(info.MarketData, info.Length)
	if err != nil {
		return nil, err
	}

	smaSrcChan := make(chan domain.SMASrc)

	go func() {
		for candle := range candleChan {
			smaSrc := domain.SMASrc{
				Value: candle.Close,
				Time:  candle.CloseTime,
			}
			smaSrcChan <- smaSrc
		}
		close(smaSrcChan)
		log.Println("candleChan closed")
	}()

	ch, err := s.smacStrategy.Subscribe(info, smaSrcChan, smaChan)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (s *StrategyService) UnsubscribeSMAC(
	info domain.SMAInfo,
	ch <-chan domain.SMACSignal,
) error {
	smaChan, ok := s.smaSignalToSMA[ch]
	if !ok {
		return fmt.Errorf("smaChan not found")
	}

	err := s.techAnalysisService.UnsubscribeSMA(info, smaChan)
	if err != nil {
		return err
	}
	delete(s.smaSignalToSMA, ch)

	candleChan, ok := s.smaSignalToCandles[ch]
	if !ok {
		return fmt.Errorf("candleChan not found")
	}

	err = s.mdService.UnsubscribeCandles(info.MarketData, candleChan)
	if err != nil {
		return err
	}
	delete(s.smaSignalToCandles, ch)

	err = s.smacStrategy.UnsubscribeSMACStrategy(info, ch)
	if err != nil {
		return err
	}

	return nil
}

func (s *StrategyService) BacktestSMAC(
	info domain.SMAInfo,
	from time.Time,
	to time.Time,
) ([]domain.SMACSignal, error) {
	smaHistory, err := s.techAnalysisService.SMAHistory(
		info,
		from,
		to,
	)
	if err != nil {
		return nil, err
	}

	candleHistory, err := s.mdService.GetCandlesByTime(
		info.MarketData,
		from,
		to,
	)
	if err != nil {
		return nil, err
	}

	srcSMA := make([]domain.SMASrc, 0, len(candleHistory))
	for _, candle := range candleHistory {
		srcSMA = append(srcSMA, domain.SMASrc{
			Value: candle.Close,
			Time:  candle.CloseTime,
		})
	}

	results, err := s.smacStrategy.Backtest(info, srcSMA, smaHistory)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *StrategyService) BacktestGoldenCross(
	info domain.GoldenCrossStrategyInfo,
	from time.Time,
	to time.Time,
) ([]domain.GoldenCrossSignal, error) {
	shortSmaHistory, err := s.techAnalysisService.SMAHistory(
		domain.SMAInfo{
			MarketData: info.Md,
			Length:     info.ShortLength,
		},
		from,
		to,
	)
	if err != nil {
		return nil, err
	}

	longSmaHistory, err := s.techAnalysisService.SMAHistory(
		domain.SMAInfo{
			MarketData: info.Md,
			Length:     info.LongLength,
		},
		from,
		to,
	)
	if err != nil {
		return nil, err
	}

	candleHistory, err := s.mdService.GetCandlesByTime(
		info.Md,
		from,
		to,
	)
	if err != nil {
		return nil, err
	}

	srcSMA := make([]domain.SMASrc, 0, len(candleHistory))
	for _, candle := range candleHistory {
		srcSMA = append(srcSMA, domain.SMASrc{
			Value: candle.Close,
			Time:  candle.CloseTime,
		})
	}

	results, err := s.goldenCrossStrategy.Backtest(
		info,
		srcSMA,
		shortSmaHistory,
		longSmaHistory,
	)
	if err != nil {
		return nil, err
	}

	return results, nil
}
