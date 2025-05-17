package tradingbots

import (
	"log"

	"github.com/Reensef/sigmasage/pkg/domain"
	"github.com/Reensef/sigmasage/pkg/exchange"
)

type SMACBot struct {
	balance        float64
	counts         map[domain.MarketData]int
	deals          []domain.SMACSignalDial
	balanceHistory []float64
	exchanger      exchange.Exchanger
	stopChan       chan struct{}
	signalChan     <-chan domain.SMACSignal
}

func NewSMACBot(
	exchanger exchange.Exchanger,
	startBalance float64,
	signalChan <-chan domain.SMACSignal,
) *SMACBot {
	return &SMACBot{
		exchanger:      exchanger,
		signalChan:     signalChan,
		balance:        startBalance,
		counts:         make(map[domain.MarketData]int),
		stopChan:       make(chan struct{}),
		balanceHistory: []float64{startBalance},
	}
}

func (s *SMACBot) Run() {
	for {
		select {
		case <-s.stopChan:
			return
		case signal := <-s.signalChan:
			s.handleSignal(signal)
		}
	}
}

func (s *SMACBot) Stop() {
	close(s.stopChan)
}

func (s *SMACBot) Deals() []domain.SMACSignalDial {
	return s.deals
}

func (s *SMACBot) Balance() float64 {
	return s.balance
}

func (s *SMACBot) BalanceHistory() []float64 {
	return s.balanceHistory
}

func (s *SMACBot) handleSignal(signal domain.SMACSignal) {
	if signal.SignalType == domain.SMACSignalType_SRC_ABOVE_SMA {
		s.handleSrcAboveSMA(signal)
	} else if signal.SignalType == domain.SMACSignalType_SRC_UNDER_SMA {
		s.handleSrcUnderSMA(signal)
	}
}

func (s *SMACBot) handleSrcAboveSMA(signal domain.SMACSignal) {
	count := s.balance / signal.LastPrice
	res, err := s.exchanger.Buy(exchange.OrderRequest{
		InstrumentID: signal.Info.MarketData.ID,
		Count:        int(count),
		Price:        signal.LastPrice,
	})
	if err != nil {
		log.Println("SMACBot: error buying", err)
		return
	}

	s.deals = append(s.deals, domain.SMACSignalDial{
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
	if _, exists := s.counts[signal.Info.MarketData]; !exists {
		s.counts[signal.Info.MarketData] = 0
	}
	s.counts[signal.Info.MarketData] += res.Count
}

func (s *SMACBot) handleSrcUnderSMA(signal domain.SMACSignal) {
	if _, exists := s.counts[signal.Info.MarketData]; !exists {
		return
	}

	res, err := s.exchanger.Sell(exchange.OrderRequest{
		InstrumentID: signal.Info.MarketData.ID,
		Count:        s.counts[signal.Info.MarketData],
		Price:        signal.LastPrice,
	})
	if err != nil {
		log.Println("SMACBot: error selling", err)
		return
	}

	s.deals = append(s.deals, domain.SMACSignalDial{
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

	s.counts[signal.Info.MarketData] -= res.Count
	if s.counts[signal.Info.MarketData] == 0 {
		delete(s.counts, signal.Info.MarketData)
	}
}
