package service

import (
	"log"
	"time"

	"github.com/Reensef/sigmasage/pkg/domain"
	"github.com/Reensef/sigmasage/pkg/techanalysis"
)

type TechAnalysisService struct {
	mdService    *MarketDataService
	smaProvider  *techanalysis.SMAProvider
	smaToCandles map[<-chan domain.SMA]<-chan domain.Candle
}

func NewTechAnalysisService(
	mdService *MarketDataService,
	smaProvider *techanalysis.SMAProvider,
) *TechAnalysisService {
	return &TechAnalysisService{
		mdService:    mdService,
		smaProvider:  smaProvider,
		smaToCandles: make(map[<-chan domain.SMA]<-chan domain.Candle),
	}
}

func (t *TechAnalysisService) SubscribeSMA(
	marketData domain.MarketData,
	length int,
) (<-chan domain.SMA, error) {
	candleHistory, err := t.mdService.GetCandlesByCount(marketData, time.Now(), length)
	if err != nil {
		return nil, err
	}

	precalcSrc := make([]float64, 0, length)
	for _, candle := range candleHistory {
		precalcSrc = append(precalcSrc, candle.Close)
	}

	candleChan, err := t.mdService.SubscribeCandles(marketData)
	if err != nil {
		return nil, err
	}

	smaSrcChan := make(chan domain.SMASrc, 100)

	smaChan, err := t.smaProvider.Subscribe(
		domain.SMAInfo{
			MarketData: marketData,
			Length:     length,
		},
		precalcSrc,
		smaSrcChan,
	)
	if err != nil {
		return nil, err
	}

	t.smaToCandles[smaChan] = candleChan

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

	return smaChan, nil
}

func (t *TechAnalysisService) UnsubscribeSMA(
	info domain.SMAInfo,
	ch <-chan domain.SMA,
) error {
	err := t.mdService.UnsubscribeCandles(info.MarketData, t.smaToCandles[ch])
	if err != nil {
		return err
	}

	err = t.smaProvider.Unsubscribe(info, ch)
	if err != nil {
		return err
	}

	delete(t.smaToCandles, ch)

	return nil
}

func (t *TechAnalysisService) SMAHistory(
	info domain.SMAInfo,
	from time.Time,
	to time.Time,
) ([]domain.SMA, error) {
	candleHistory, err := t.mdService.GetCandlesByCount(info.MarketData, from, info.Length)
	if err != nil {
		return nil, err
	}

	precalcSrc := make([]float64, 0, len(candleHistory))
	for _, candle := range candleHistory {
		precalcSrc = append(precalcSrc, candle.Close)
	}

	candleHistory, err = t.mdService.GetCandlesByTime(info.MarketData, from, to)
	if err != nil {
		return nil, err
	}

	src := make([]domain.SMASrc, 0, len(candleHistory))
	for _, candle := range candleHistory {
		src = append(src, domain.SMASrc{Value: candle.Close, Time: candle.CloseTime})
	}

	smaResults, err := t.smaProvider.CalcFromSrc(info, precalcSrc, src)
	if err != nil {
		return nil, err
	}

	return smaResults, nil
}
