package domain

import (
	"time"
)

// Сигнал стратегии пересечения скользящих средних (SMA Crossover)
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

type GoldenCrossStrategyInfo struct {
	Md          MarketData
	ShortLength int
	LongLength  int
}

type GoldenCrossSignalType int

const (
	GoldenCrossSignalType_GOLDEN_CROSS GoldenCrossSignalType = iota
	GoldenCrossSignalType_DEATH_CROSS
	GoldenCrossSignalType_NO_CROSS
)

type GoldenCrossSignal struct {
	Md         MarketData
	LastPrice  float64
	SignalType GoldenCrossSignalType
	Time       time.Time
}
