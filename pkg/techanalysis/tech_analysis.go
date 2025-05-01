package techanalysis

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/Reensef/sigmasage/pkg/marketdata"
)

type SMAInfo struct {
	marketData marketdata.MarketData
	length     int
}
type SMA struct {
	Info  SMAInfo
	Value float64
	Time  time.Time
}

type TechAnalysisService struct {
	mdService        marketdata.MarketDataServicer
	smaSubscribers   map[SMAInfo][]chan SMA
	smaStreamCancels map[SMAInfo]context.CancelFunc
}

func NewTechAnalysisService(
	mdService marketdata.MarketDataServicer,
) (*TechAnalysisService, error) {
	return &TechAnalysisService{
		mdService: mdService,
	}, nil
}

func (t *TechAnalysisService) SubscribeSMA(info SMAInfo) (<-chan SMA, error) {
	ch := make(chan SMA, 100)

	needStartNotify := false

	if _, exists := t.smaSubscribers[info]; !exists {
		t.smaSubscribers[info] = make([]chan SMA, 0)
		needStartNotify = true
	}
	t.smaSubscribers[info] = append(t.smaSubscribers[info], ch)

	if needStartNotify {
		ctx, cancel := context.WithCancel(context.Background())
		t.smaStreamCancels[info] = cancel
		go t.startCalculateSMA(info, ctx)
	}

	return ch, nil
}

func (t *TechAnalysisService) UnsubscribeSMA(info SMAInfo, ch <-chan SMA) error {
	if _, exists := t.smaSubscribers[info]; !exists {
		return fmt.Errorf("undefined subscriber")
	}

	for i, c := range t.smaSubscribers[info] {
		if c == ch {
			t.smaSubscribers[info] = slices.Delete(t.smaSubscribers[info], i, i+1)

			if len(t.smaSubscribers[info]) == 0 {
				t.smaStreamCancels[info]()
				delete(t.smaStreamCancels, info)
				delete(t.smaSubscribers, info)
			}
			return nil
		}
	}

	return fmt.Errorf("undefined subscriber")
}

func (t *TechAnalysisService) GetSMAHistory(info SMAInfo, from time.Time, to time.Time) ([]SMA, error) {
	timeFrame := marketdata.ConvertMarketDataIntervalToTime(info.marketData.Interval)

	if !from.Truncate(timeFrame).Equal(from) || !to.Truncate(timeFrame).Equal(to) {
		return nil, fmt.Errorf("from and to must be aligned to marketdata interval: %s", timeFrame.String())
	}

	historyData, err := t.mdService.GetCandles(
		info.marketData,
		from.Add(-time.Duration(info.length)*timeFrame),
		to,
	)

	if err != nil {
		return nil, err
	}

	if len(historyData) < info.length+int(to.Sub(from)/timeFrame) {
		return nil, fmt.Errorf("not enough data to calculate SMA")
	}

	smaSrc := make([]float64, 0)
	for _, candle := range historyData {
		smaSrc = append(smaSrc, candle.Close)
	}

	smaCalculator, err := NewSMACalculator(info.length, smaSrc[:info.length])
	if err != nil {
		return nil, err
	}

	smaResults := make([]SMA, 0)

	for _, candle := range historyData[info.length:] {
		sma := smaCalculator.Update(candle.Close)
		smaResults = append(smaResults, SMA{
			Info:  info,
			Value: sma,
			Time:  candle.Time,
		})
	}

	return smaResults, nil
}

func (t *TechAnalysisService) startCalculateSMA(info SMAInfo, ctx context.Context) {
	candlesChan, err := t.mdService.SubscribeCandles(marketdata.MarketData{
		ID:       info.marketData.ID,
		Interval: info.marketData.Interval,
	})

	if err != nil {
		log.Println("Error subscribing to candles:", err)
		return
	}

	duration := time.Duration(info.length) * marketdata.ConvertMarketDataIntervalToTime(info.marketData.Interval)
	historyData, err := t.mdService.GetCandles(
		info.marketData,
		time.Now().Add(-duration),
		time.Now(),
	)
	if err != nil {
		log.Println("Error getting historical candles:", err)
		return
	}

	smaSrc := make([]float64, 0)

	for _, candle := range historyData {
		smaSrc = append(smaSrc, candle.Close)
	}

	smaCalculator, err := NewSMACalculator(info.length, smaSrc)
	if err != nil {
		log.Println("Error creating SMA calculator:", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case candle, ok := <-candlesChan:
			if !ok {
				log.Println("Error receiving candle for SMA calculation")
				return
			}

			sma := smaCalculator.Update(candle.Close)

			for _, subscriber := range t.smaSubscribers[info] {
				subscriber <- SMA{
					Info:  info,
					Value: sma,
					Time:  candle.Time,
				}
			}
		}
	}
}
