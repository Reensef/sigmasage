package domain

import "time"

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
	OpenTime   time.Time
	CloseTime  time.Time
}

type MarketDataProviderType int32

const (
	MarketDataProviderType_TINKOFF MarketDataProviderType = iota
)

type MarketDataInterval int32

const (
	MarketDataInterval_UNSPECIFIED     MarketDataInterval = iota
	MarketDataInterval_ONE_MINUTE      MarketDataInterval = iota
	MarketDataInterval_TWO_MIN         MarketDataInterval = iota
	MarketDataInterval_THREE_MIN       MarketDataInterval = iota
	MarketDataInterval_FIVE_MINUTES    MarketDataInterval = iota
	MarketDataInterval_TEN_MIN         MarketDataInterval = iota
	MarketDataInterval_FIFTEEN_MINUTES MarketDataInterval = iota
	MarketDataInterval_THERTY_MIN      MarketDataInterval = iota
	MarketDataInterval_ONE_HOUR        MarketDataInterval = iota
	MarketDataInterval_TWO_HOUR        MarketDataInterval = iota
	MarketDataInterval_FOUR_HOUR       MarketDataInterval = iota
	MarketDataInterval_ONE_DAY         MarketDataInterval = iota
	MarketDataInterval_WEEK            MarketDataInterval = iota
	MarketDataInterval_MONTH           MarketDataInterval = iota
)
