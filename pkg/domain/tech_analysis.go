package domain

import (
	"time"
)

type SMAInfo struct {
	MarketData MarketData
	Length     int
}

type SMA struct {
	Info  SMAInfo
	Value float64
	Time  time.Time
}

type SMASrc struct {
	Value float64
	Time  time.Time
}

// Сигнал стратегии пересечения скользящих средних (SimpleMovingAvgCross)
type SMACSignalType int

const (
	SMACSignalType_SRC_ABOVE_SMA SMACSignalType = iota
	SMACSignalType_SRC_UNDER_SMA
	SMACSignalType_NO_CROSS
)

type SMACSignal struct {
	Info       SMAInfo
	LastPrice  float64
	SignalType SMACSignalType
	Time       time.Time
}
