package strategy

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Reensef/sigmasage/pkg/domain"
)

type SMACStatus struct {
	isSrcAboveSMA bool
	lastSrc       domain.SMASrc
	lastSMA       domain.SMA
}

type SMACStrategy struct {
	subscribers map[domain.SMAInfo][]chan domain.SMACSignal
	status      map[domain.SMAInfo]*SMACStatus
	runners     map[domain.SMAInfo]context.CancelFunc
}

func NewSMACStrategy() *SMACStrategy {
	return &SMACStrategy{
		subscribers: make(map[domain.SMAInfo][]chan domain.SMACSignal),
		runners:     make(map[domain.SMAInfo]context.CancelFunc),
	}
}

func (s *SMACStrategy) Subscribe(
	config domain.SMAInfo,
	srcChan <-chan domain.SMASrc,
	smaChan <-chan domain.SMA,
) (<-chan domain.SMACSignal, error) {
	ch := make(chan domain.SMACSignal, 100)

	if _, exists := s.subscribers[config]; !exists {
		s.subscribers[config] = make([]chan domain.SMACSignal, 0)
		ctx, cancel := context.WithCancel(context.Background())

		s.runners[config] = cancel
		firstSrc, firstSMA := s.syncSMACData(srcChan, smaChan, ctx)

		s.status[config] = &SMACStatus{
			isSrcAboveSMA: firstSrc.Value > firstSMA.Value,
			lastSrc:       firstSrc,
			lastSMA:       firstSMA,
		}

		s.run(config, srcChan, smaChan, ctx)
	}

	s.subscribers[config] = append(s.subscribers[config], ch)

	return ch, nil
}

func (s *SMACStrategy) UnsubscribeSMACStrategy(
	config domain.SMAInfo,
	ch <-chan domain.SMACSignal,
) error {
	ok := false

	if subscribers, exists := s.subscribers[config]; exists {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				s.subscribers[config] = slices.Delete(subscribers, i, i+1)
				close(subscriber)
				ok = true
				break
			}
		}

		if len(s.subscribers[config]) == 0 {
			delete(s.subscribers, config)
			s.runners[config]()
			delete(s.runners, config)
			delete(s.status, config)
		}
	}

	if !ok {
		return fmt.Errorf("subscribe not found")
	}

	return nil
}

func (s *SMACStrategy) run(
	config domain.SMAInfo,
	srcChan <-chan domain.SMASrc,
	smaChan <-chan domain.SMA,
	ctx context.Context,
) {
	for {
		status, exists := s.status[config]
		if !exists {
			return
		}

		select {
		case <-ctx.Done():
			return
		case src := <-srcChan:
			status.lastSrc = src

			if time.Until(src.Time).Abs() < time.Second {
				if status.lastSrc.Time == status.lastSMA.Time {
					signal := s.makeDecision(status)
					s.notifySMACSubscribers(config, signal, status.lastSrc.Time)
				}
			}
		case sma := <-smaChan:
			status.lastSMA = sma

			if time.Until(status.lastSMA.Time).Abs() < time.Second {
				if status.lastSrc.Time == status.lastSMA.Time {
					signal := s.makeDecision(status)
					s.notifySMACSubscribers(config, signal, status.lastSrc.Time)
				}
			}
		}
		s.status[config] = status
	}
}

// TODO пересмотреть синхронизацию данных
// В синнхронизации нет необходимости, если время не совпадет это признак ошибки
func (s *SMACStrategy) syncSMACData(
	candlesChan <-chan domain.SMASrc,
	smaChan <-chan domain.SMA,
	ctx context.Context,
) (domain.SMASrc, domain.SMA) {
	var src *domain.SMASrc
	var sma *domain.SMA

	for {
		select {
		case <-ctx.Done():
			return domain.SMASrc{}, domain.SMA{}
		case recvCandle := <-candlesChan:
			src = &recvCandle

			if sma != nil && src.Time == sma.Time {
				return *src, *sma
			}
		case recvSMA := <-smaChan:
			sma = &recvSMA
			if src != nil && src.Time == sma.Time {
				return *src, *sma
			}
		}
	}
}

func (s *SMACStrategy) makeDecision(status *SMACStatus) domain.SMACSignalType {
	if status.lastSrc.Value > status.lastSMA.Value {
		if !status.isSrcAboveSMA {
			status.isSrcAboveSMA = true
			return domain.SMACSignalType_SRC_ABOVE_SMA
		}

		return domain.SMACSignalType_NO_CROSS
	} else if status.lastSrc.Value < status.lastSMA.Value {
		if status.isSrcAboveSMA {
			status.isSrcAboveSMA = false
			return domain.SMACSignalType_SRC_UNDER_SMA
		}

		return domain.SMACSignalType_NO_CROSS
	}

	return domain.SMACSignalType_NO_CROSS
}

func (s *SMACStrategy) notifySMACSubscribers(info domain.SMAInfo, signalType domain.SMACSignalType, time time.Time) {
	subscribers, exists := s.subscribers[info]
	if !exists {
		return
	}

	if signalType == domain.SMACSignalType_NO_CROSS {
		return
	}

	signal := domain.SMACSignal{
		Info:       info,
		SignalType: signalType,
		Time:       time,
	}

	for _, subscriber := range subscribers {
		subscriber <- signal
	}
}

func (s *SMACStrategy) Backtest(
	info domain.SMAInfo,
	src []domain.SMASrc,
	sma []domain.SMA,
) ([]domain.SMACSignal, error) {
	if len(src) != len(sma) {
		return nil, fmt.Errorf("src and sma have different length")
	}

	signals := make([]domain.SMACSignal, 0)

	status := SMACStatus{
		isSrcAboveSMA: src[0].Value > sma[0].Value,
		lastSrc:       src[0],
		lastSMA:       sma[0],
	}

	for i := 1; i < len(src); i++ {
		status.lastSrc = src[i]
		status.lastSMA = sma[i]

		signal := s.makeDecision(&status)
		if signal != domain.SMACSignalType_NO_CROSS {
			signals = append(signals, domain.SMACSignal{
				Info:       info,
				SignalType: signal,
				Time:       status.lastSrc.Time,
				LastPrice:  status.lastSrc.Value,
			})
		}
	}

	return signals, nil
}
