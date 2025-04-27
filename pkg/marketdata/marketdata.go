package marketdata

import "time"

type MarketDataInfo struct {
	IntrumentID string
	Interval    MarketDataInterval
}

type Candle struct {
	MarketDataInfo MarketDataInfo
	Open           float64
	High           float64
	Low            float64
	Close          float64
	Volume         float64
	Time           time.Time
}

type CandleData interface {
	SubscribeCandles(marketDataInfo MarketDataInfo) (<-chan Candle, error)
	UnsubscribeCandles(marketDataInfo MarketDataInfo, ch <-chan Candle) bool
	GetCandles(marketDataInfo MarketDataInfo, from time.Time, to time.Time) ([]Candle, error)
}

type MarketDataInterval int32

const (
	UNSPECIFIED     MarketDataInterval = iota
	ONE_MINUTE      MarketDataInterval = iota
	TWO_MIN         MarketDataInterval = iota
	THREE_MIN       MarketDataInterval = iota
	FIVE_MINUTES    MarketDataInterval = iota
	TEN_MIN         MarketDataInterval = iota
	FIFTEEN_MINUTES MarketDataInterval = iota
	THERTY_MIN      MarketDataInterval = iota
	ONE_HOUR        MarketDataInterval = iota
	TWO_HOUR        MarketDataInterval = iota
	FOUR_HOUR       MarketDataInterval = iota
	ONE_DAY         MarketDataInterval = iota
	WEEK            MarketDataInterval = iota
	MONTH           MarketDataInterval = iota
)

func ConvertMarketDataIntervalToTime(interval MarketDataInterval) time.Duration {
	switch interval {
	case ONE_MINUTE:
		return time.Minute
	case TWO_MIN:
		return 2 * time.Minute
	case THREE_MIN:
		return 3 * time.Minute
	case FIVE_MINUTES:
		return 5 * time.Minute
	case TEN_MIN:
		return 10 * time.Minute
	case FIFTEEN_MINUTES:
		return 15 * time.Minute
	case THERTY_MIN:
		return 30 * time.Minute
	case ONE_HOUR:
		return time.Hour
	case TWO_HOUR:
		return 2 * time.Hour
	case FOUR_HOUR:
		return 4 * time.Hour
	case ONE_DAY:
		return 24 * time.Hour
	case WEEK:
		return 7 * 24 * time.Hour
	case MONTH:
		return 30 * 24 * time.Hour
	default:
		return 0
	}
}
