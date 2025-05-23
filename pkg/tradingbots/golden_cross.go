package tradingbots

import (
	"log"

	"github.com/Reensef/sigmasage/pkg/domain"
	"github.com/Reensef/sigmasage/pkg/exchange"
)

// TODO Наследование?
type GoldenCrossBot struct {
	balance        float64
	counts         map[domain.MarketData]int
	deals          []domain.GoldenCrossSignalDial
	balanceHistory []float64
	exchanger      exchange.Exchanger
	stopChan       chan struct{}
	signalChan     <-chan domain.GoldenCrossSignal
}

func NewGoldenCrossBot(
	exchanger exchange.Exchanger,
	startBalance float64,
	signalChan <-chan domain.GoldenCrossSignal,
) *GoldenCrossBot {
	return &GoldenCrossBot{
		exchanger:      exchanger,
		signalChan:     signalChan,
		balance:        startBalance,
		counts:         make(map[domain.MarketData]int),
		stopChan:       make(chan struct{}),
		balanceHistory: []float64{startBalance},
	}
}

func (s *GoldenCrossBot) Run() {
	for {
		select {
		case <-s.stopChan:
			return
		case signal := <-s.signalChan:
			s.handleSignal(signal)
		}
	}
}

func (s *GoldenCrossBot) Stop() {
	close(s.stopChan)
}

func (s *GoldenCrossBot) Deals() []domain.GoldenCrossSignalDial {
	return s.deals
}

func (s *GoldenCrossBot) Balance() float64 {
	return s.balance
}

func (s *GoldenCrossBot) BalanceHistory() []float64 {
	return s.balanceHistory
}

func (s *GoldenCrossBot) handleSignal(signal domain.GoldenCrossSignal) {
	if signal.SignalType == domain.GoldenCrossSignalType_GOLDEN_CROSS {
		s.handleGoldenCross(signal)
	} else if signal.SignalType == domain.GoldenCrossSignalType_DEATH_CROSS {
		s.handleDeathCross(signal)
	}
}

func (s *GoldenCrossBot) handleGoldenCross(signal domain.GoldenCrossSignal) {
	count := s.balance / signal.LastPrice
	res, err := s.exchanger.Buy(exchange.OrderRequest{
		InstrumentID: signal.Md.ID,
		Count:        int(count),
		Price:        signal.LastPrice,
	})
	if err != nil {
		log.Println("SMACBot: error buying", err)
		return
	}

	s.deals = append(s.deals, domain.GoldenCrossSignalDial{
		Deal: domain.Deal{
			Direction:  domain.DealDirection_BUY,
			Time:       res.Time,
			Price:      signal.LastPrice,
			Count:      res.Count,
			LotPrice:   res.LotPrice,
			Commission: res.Commission,
		},
		Signal: signal,
	})

	s.balance -= res.LotPrice
	s.balanceHistory = append(s.balanceHistory, s.balance)
	if _, exists := s.counts[signal.Md]; !exists {
		s.counts[signal.Md] = 0
	}
	s.counts[signal.Md] += res.Count
}

func (s *GoldenCrossBot) handleDeathCross(signal domain.GoldenCrossSignal) {
	if _, exists := s.counts[signal.Md]; !exists {
		return
	}

	res, err := s.exchanger.Sell(exchange.OrderRequest{
		InstrumentID: signal.Md.ID,
		Count:        s.counts[signal.Md],
		Price:        signal.LastPrice,
	})
	if err != nil {
		log.Println("SMACBot: error selling", err)
		return
	}

	s.deals = append(s.deals, domain.GoldenCrossSignalDial{
		Deal: domain.Deal{
			Direction:  domain.DealDirection_SELL,
			Time:       res.Time,
			Price:      signal.LastPrice,
			Count:      res.Count,
			LotPrice:   res.LotPrice,
			Commission: res.Commission,
		},
		Signal: signal,
	})

	s.balance += res.LotPrice
	s.balanceHistory = append(s.balanceHistory, s.balance)

	s.counts[signal.Md] -= res.Count
	if s.counts[signal.Md] == 0 {
		delete(s.counts, signal.Md)
	}
}
