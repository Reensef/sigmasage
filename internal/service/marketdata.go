package service

import (
	"fmt"
	"time"

	"github.com/Reensef/sigmasage/pkg/domain"
	"github.com/Reensef/sigmasage/pkg/marketdata"
)

type MarketDataService struct {
	tinkoffProvider marketdata.MarketDataProvider
}

func NewMarketDataService(tinkoffProvider marketdata.MarketDataProvider) (
	*MarketDataService,
	error,
) {
	return &MarketDataService{
		tinkoffProvider: tinkoffProvider,
	}, nil
}

func (m *MarketDataService) SubscribeCandles(marketData domain.MarketData) (
	<-chan domain.Candle,
	error,
) {
	if marketData.ProviderType == domain.MarketDataProviderType_TINKOFF {
		resultChan, err := m.tinkoffProvider.SubscribeCandles(marketData)
		if err != nil {
			return nil, err
		}
		return resultChan, nil
	} else {
		return nil, fmt.Errorf("undefined provider type")
	}
}

func (m *MarketDataService) UnsubscribeCandles(
	marketData domain.MarketData,
	ch <-chan domain.Candle,
) error {
	if marketData.ProviderType == domain.MarketDataProviderType_TINKOFF {
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
	marketData domain.MarketData,
	from time.Time,
	to time.Time,
) ([]domain.Candle, error) {
	if marketData.ProviderType == domain.MarketDataProviderType_TINKOFF {
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
	marketData domain.MarketData,
	to time.Time,
	count int,
) ([]domain.Candle, error) {
	if marketData.ProviderType == domain.MarketDataProviderType_TINKOFF {
		candles, err := m.tinkoffProvider.GetCandlesByCount(marketData, to, count)
		if err != nil {
			return nil, err
		}
		return candles, nil
	} else {
		return nil, fmt.Errorf("undefined provider type")
	}
}
