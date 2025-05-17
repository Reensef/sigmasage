package marketdata

import (
	"time"

	"github.com/Reensef/sigmasage/pkg/domain"
)

type MarketDataProvider interface {
	SubscribeCandles(marketDataInfo domain.MarketData) (<-chan domain.Candle, error)
	UnsubscribeCandles(marketDataInfo domain.MarketData, ch <-chan domain.Candle) error
	GetCandlesByCount(marketDataInfo domain.MarketData, to time.Time, count int) ([]domain.Candle, error)
	GetCandlesByTime(marketDataInfo domain.MarketData, from time.Time, to time.Time) ([]domain.Candle, error)
}
