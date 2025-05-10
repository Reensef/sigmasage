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
	MarketData marketdata.MarketData
	Length     int
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
		mdService:        mdService,
		smaSubscribers:   make(map[SMAInfo][]chan SMA),
		smaStreamCancels: make(map[SMAInfo]context.CancelFunc),
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
		go t.startCalculatingSMA(info, ctx)
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

func (t *TechAnalysisService) SMAHistory(info SMAInfo, from time.Time, to time.Time) ([]SMA, error) {
	if to.Before(from) {
		return nil, fmt.Errorf("to must be greater than from")
	}

	timeFrame := marketdata.ConvertMarketDataIntervalToTime(info.MarketData.Interval)

	if !from.Truncate(timeFrame).Equal(from) || !to.Truncate(timeFrame).Equal(to) {
		return nil, fmt.Errorf("from and to must be aligned to marketdata interval: %s", timeFrame.String())
	}

	historyDataForPrecalc, err := t.mdService.GetCandlesByCount(
		info.MarketData,
		from,
		50,
	)

	if len(historyDataForPrecalc) != info.Length {
		return nil, fmt.Errorf("not enough data to calculate SMA")
	}

	if err != nil {
		return nil, err
	}

	historyDataForCalc, err := t.mdService.GetCandlesByTime(
		info.MarketData,
		from,
		to,
	)

	if err != nil {
		return nil, err
	}

	smaPrecalcSrc := make([]float64, len(historyDataForPrecalc))
	for i, candle := range historyDataForPrecalc {
		smaPrecalcSrc[i] = candle.Close
	}

	smaSrc := make([]float64, len(historyDataForCalc))
	for i, candle := range historyDataForCalc {
		smaSrc[i] = candle.Close
	}

	smaCalculator, err := NewSMACalculator(info.Length, smaPrecalcSrc)
	if err != nil {
		return nil, err
	}

	smaResults := make([]SMA, 0)

	for _, candle := range historyDataForCalc {
		sma := smaCalculator.Update(candle.Close)
		smaResults = append(smaResults, SMA{
			Info:  info,
			Value: sma,
			Time:  candle.StartTime,
		})
	}

	return smaResults, nil
}

func (t *TechAnalysisService) startCalculatingSMA(info SMAInfo, ctx context.Context) {
	candlesChan, err := t.mdService.SubscribeCandles(info.MarketData)

	if err != nil {
		log.Println("Error subscribing to candles:", err)
		return
	}

	duration := time.Duration(info.Length) * marketdata.ConvertMarketDataIntervalToTime(info.MarketData.Interval)
	historyData, err := t.mdService.GetCandlesByTime(
		info.MarketData,
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

	smaCalculator, err := NewSMACalculator(info.Length, smaSrc)
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
					Time:  candle.EndTime,
				}
			}
		}
	}
}
