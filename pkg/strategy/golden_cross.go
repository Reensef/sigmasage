package strategy

import (
	"context"
	"fmt"

	"github.com/Reensef/sigmasage/pkg/domain"
)

type GoldenCrossStatus struct {
	isShortAboveLong bool
	lastShortSMA     domain.SMA
	lastLongSMA      domain.SMA
}

type GoldenCrossStrategy struct {
	subscribers map[domain.GoldenCrossStrategyInfo][]chan domain.GoldenCrossSignal
	// status      map[domain.GoldenCrossStrategyInfo]*GoldenCrossStatus
	runners map[domain.GoldenCrossStrategyInfo]context.CancelFunc
}

func NewGoldenCrossStrategy() *GoldenCrossStrategy {
	return &GoldenCrossStrategy{
		subscribers: make(map[domain.GoldenCrossStrategyInfo][]chan domain.GoldenCrossSignal),
		runners:     make(map[domain.GoldenCrossStrategyInfo]context.CancelFunc),
	}
}

func (s *GoldenCrossStrategy) makeDecision(status *GoldenCrossStatus) domain.GoldenCrossSignalType {
	if status.lastShortSMA.Value > status.lastLongSMA.Value {
		if !status.isShortAboveLong {
			status.isShortAboveLong = true
			return domain.GoldenCrossSignalType_GOLDEN_CROSS
		}

		return domain.GoldenCrossSignalType_NO_CROSS
	} else if status.lastShortSMA.Value < status.lastLongSMA.Value {
		if status.isShortAboveLong {
			status.isShortAboveLong = false
			return domain.GoldenCrossSignalType_DEATH_CROSS
		}

		return domain.GoldenCrossSignalType_NO_CROSS
	}

	return domain.GoldenCrossSignalType_NO_CROSS
}

func (s *GoldenCrossStrategy) Backtest(
	info domain.GoldenCrossStrategyInfo,
	src []domain.SMASrc,
	shortSMA []domain.SMA,
	longSMA []domain.SMA,
) ([]domain.GoldenCrossSignal, error) {
	if len(src) != len(shortSMA) || len(src) != len(longSMA) {
		return nil, fmt.Errorf("src, short SMA, and long SMA must have the same length")
	}

	if len(shortSMA) < 1 {
		return nil, fmt.Errorf("not enough data")
	}

	signals := make([]domain.GoldenCrossSignal, 0)

	status := GoldenCrossStatus{
		isShortAboveLong: shortSMA[0].Value > longSMA[0].Value,
		lastShortSMA:     shortSMA[0],
		lastLongSMA:      longSMA[0],
	}

	for i := 1; i < len(shortSMA); i++ {
		status.lastShortSMA = shortSMA[i]
		status.lastLongSMA = longSMA[i]

		signal := s.makeDecision(&status)
		if signal != domain.GoldenCrossSignalType_NO_CROSS {
			signals = append(signals, domain.GoldenCrossSignal{
				LastPrice:  src[i].Value,
				SignalType: signal,
				Time:       src[i].Time,
			})
		}
	}

	return signals, nil
}
