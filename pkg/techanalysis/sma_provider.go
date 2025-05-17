package techanalysis

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/Reensef/sigmasage/pkg/domain"
)

type SMAProvider struct {
	subscribers   map[domain.SMAInfo][]chan domain.SMA
	streamCancels map[domain.SMAInfo]context.CancelFunc
}

func NewSMAProvider() *SMAProvider {
	return &SMAProvider{
		subscribers:   make(map[domain.SMAInfo][]chan domain.SMA),
		streamCancels: make(map[domain.SMAInfo]context.CancelFunc),
	}
}

func (s *SMAProvider) Subscribe(
	info domain.SMAInfo,
	src []float64,
	srcCh chan domain.SMASrc,
) (<-chan domain.SMA, error) {
	ch := make(chan domain.SMA, 100)

	needStartNotify := false

	if _, exists := s.subscribers[info]; !exists {
		s.subscribers[info] = make([]chan domain.SMA, 0)
		needStartNotify = true
	}
	s.subscribers[info] = append(s.subscribers[info], ch)

	if needStartNotify {
		ctx, cancel := context.WithCancel(context.Background())
		s.streamCancels[info] = cancel
		go s.startStream(info, src, srcCh, ctx)
	}

	return ch, nil
}

func (s *SMAProvider) Unsubscribe(info domain.SMAInfo, ch <-chan domain.SMA) error {
	if _, exists := s.subscribers[info]; !exists {
		return fmt.Errorf("undefined subscriber")
	}

	for i, c := range s.subscribers[info] {
		if c == ch {
			s.subscribers[info] = slices.Delete(s.subscribers[info], i, i+1)

			if len(s.subscribers[info]) == 0 {
				s.streamCancels[info]()
				delete(s.streamCancels, info)
				delete(s.subscribers, info)
			}
			return nil
		}
	}

	return fmt.Errorf("undefined subscriber")
}

func (s *SMAProvider) CalcFromSrc(info domain.SMAInfo, precalcSrc []float64, src []domain.SMASrc) ([]domain.SMA, error) {
	smaCalculator, err := NewSMACalculator(info.Length, precalcSrc)
	if err != nil {
		return nil, err
	}

	smaResults := make([]domain.SMA, 0)

	for _, s := range src {
		sma := smaCalculator.Update(s.Value)
		smaResults = append(smaResults, domain.SMA{
			Info:  info,
			Value: sma,
			Time:  s.Time,
		})
	}

	return smaResults, nil
}

func (s *SMAProvider) startStream(info domain.SMAInfo, src []float64, srcCh chan domain.SMASrc, ctx context.Context) {
	smaCalculator, err := NewSMACalculator(info.Length, src)
	if err != nil {
		log.Println("Error creating SMA calculator:", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case src, ok := <-srcCh:
			if !ok {
				log.Println("Error receiving candle for SMA calculation")
				return
			}

			sma := smaCalculator.Update(src.Value)

			for _, subscriber := range s.subscribers[info] {
				subscriber <- domain.SMA{
					Info:  info,
					Value: sma,
					Time:  src.Time,
				}
			}
		}
	}
}
