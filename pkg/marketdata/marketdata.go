package marketdata

import (
	"fmt"
	"time"
)

type MarketData struct {
	ID           string
	Interval     MarketDataInterval
	ProviderType MarketDataProviderType
}

type Candle struct {
	MarketData MarketData
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
	StartTime  time.Time // Open time
	EndTime    time.Time // Close time
}

type MarketDataProvider interface {
	SubscribeCandles(marketDataInfo MarketData) (<-chan Candle, error)
	UnsubscribeCandles(marketDataInfo MarketData, ch <-chan Candle) error
	GetCandlesByCount(marketDataInfo MarketData, to time.Time, count int) ([]Candle, error)
	GetCandlesByTime(marketDataInfo MarketData, from time.Time, to time.Time) ([]Candle, error)
}

type MarketDataProviderType int32

const (
	TINKOFF MarketDataProviderType = iota
)

type MarketDataServicer interface {
	SubscribeCandles(marketData MarketData) (<-chan Candle, error)
	UnsubscribeCandles(marketData MarketData, ch <-chan Candle) error
	GetCandlesByCount(marketData MarketData, to time.Time, count int) ([]Candle, error)
	GetCandlesByTime(marketData MarketData, from time.Time, to time.Time) ([]Candle, error)
}

type MarketDataService struct {
	tinkoffToken    string
	tinkoffProvider MarketDataProvider
}

func NewMarketDataService(
	tinkoffToken string,
) (*MarketDataService, error) {
	tinkoffMD, err := NewTinkoffMarketdata(tinkoffToken)
	if err != nil {
		return nil, err
	}

	return &MarketDataService{
		tinkoffToken:    tinkoffToken,
		tinkoffProvider: tinkoffMD,
	}, nil
}

func (m *MarketDataService) SubscribeCandles(marketData MarketData) (<-chan Candle, error) {
	if marketData.ProviderType == TINKOFF {
		resultChan, err := m.tinkoffProvider.SubscribeCandles(marketData)
		if err != nil {
			return nil, err
		}
		return resultChan, nil
	} else {
		return nil, fmt.Errorf("undefined provider type")
	}
}

func (m *MarketDataService) UnsubscribeCandles(marketData MarketData, ch <-chan Candle) error {
	if marketData.ProviderType == TINKOFF {
		err := m.tinkoffProvider.UnsubscribeCandles(marketData, ch)
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("undefined provider type")
	}
}

func (m *MarketDataService) GetCandlesByTime(
	marketData MarketData,
	from time.Time,
	to time.Time,
) ([]Candle, error) {
	if marketData.ProviderType == TINKOFF {
		candles, err := m.tinkoffProvider.GetCandlesByTime(marketData, from, to)
		if err != nil {
			return nil, err
		}
		return candles, nil
	} else {
		return nil, fmt.Errorf("undefined provider type")
	}
}

func (m *MarketDataService) GetCandlesByCount(
	marketData MarketData,
	to time.Time,
	count int,
) ([]Candle, error) {
	if marketData.ProviderType == TINKOFF {
		candles, err := m.tinkoffProvider.GetCandlesByCount(marketData, to, count)
		if err != nil {
			return nil, err
		}
		return candles, nil
	} else {
		return nil, fmt.Errorf("undefined provider type")
	}
}
