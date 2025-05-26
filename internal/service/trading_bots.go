package service

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Reensef/sigmasage/pkg/domain"
	"github.com/Reensef/sigmasage/pkg/exchange"
	"github.com/Reensef/sigmasage/pkg/tradingbots"
)

type TradingBotService struct {
	mdService       *MarketDataService
	strategyService *StrategyService
	smacBots        map[int64]*tradingbots.SMACBot // TODO Наследование?
	smacBotInfo     map[int64]domain.SMAInfo
	smacSignals     map[domain.SMAInfo]<-chan domain.SMACSignal
}

func NewTradingBotService(strategyService *StrategyService, mdService *MarketDataService) *TradingBotService {
	return &TradingBotService{strategyService: strategyService, mdService: mdService}
}

func (t *TradingBotService) CreateSMACBot(
	info domain.SMAInfo,
	exchanger exchange.Exchanger,
	startBalance float64,
) (int64, error) {
	signalChan, err := t.strategyService.SubscribeSMAC(info)
	if err != nil {
		return 0, err
	}

	bot := tradingbots.NewSMACBot(exchanger, startBalance, signalChan)

	t.smacBots[0] = bot
	t.smacBotInfo[0] = info
	t.smacSignals[info] = signalChan

	return 0, nil
}

func (t *TradingBotService) DeleteSMACBot(id int64) error {
	bot, ok := t.smacBots[0]
	if !ok {
		return fmt.Errorf("bot not found")
	}

	bot.Stop()
	delete(t.smacBots, 0)

	info, exists := t.smacBotInfo[0]
	if !exists {
		return fmt.Errorf("bot info not found")
	}

	signalChan, exists := t.smacSignals[info]
	if !exists {
		return fmt.Errorf("signal channel not found")
	}

	t.strategyService.UnsubscribeSMAC(info, signalChan)

	delete(t.smacSignals, info)
	delete(t.smacBotInfo, 0)

	return nil
}

func (t *TradingBotService) RunSMACBot(id int64) error {
	bot, ok := t.smacBots[id]
	if !ok {
		return fmt.Errorf("bot not found")
	}

	go bot.Run()
	return nil
}

func (t *TradingBotService) StopSMACBot(id int64) error {
	bot, ok := t.smacBots[id]
	if !ok {
		return fmt.Errorf("bot not found")
	}

	bot.Stop()
	return nil
}

func (t *TradingBotService) BacktestSMAC(
	smaInfo domain.SMAInfo,
	startBalance float64,
	exchanger exchange.Exchanger,
	from time.Time,
	to time.Time,
) (resultSignalDeals []domain.SMACSignalDial, balanceHistory []float64, err error) {
	signals, err := t.strategyService.BacktestSMAC(smaInfo, from, to)
	if err != nil {
		return nil, nil, err
	}

	signalChan := make(chan domain.SMACSignal)
	bot := tradingbots.NewSMACBot(exchanger, startBalance, signalChan)

	go bot.Run()

	for _, signal := range signals {
		signalChan <- signal
	}
	bot.Stop()
	close(signalChan)

	return bot.Deals(), bot.BalanceHistory(), nil
}

func (t *TradingBotService) BacktestGoldenCross(
	strategyInfo domain.GoldenCrossStrategyInfo,
	startBalance float64,
	commissionPercent float64,
	slippagePercent float64,
	from time.Time,
	to time.Time,
) (resultSignalDeals []domain.GoldenCrossSignalDial, balanceHistory []float64, err error) {
	signals, err := t.strategyService.BacktestGoldenCross(strategyInfo, from, to)
	if err != nil {
		return nil, nil, err
	}

	exchanger := exchange.NewMockExchange(
		commissionPercent,
		slippagePercent,
	)

	signalChan := make(chan domain.GoldenCrossSignal)
	bot := tradingbots.NewGoldenCrossBot(exchanger, startBalance, signalChan)

	go bot.Run()

	for _, signal := range signals {
		signalChan <- signal
	}
	bot.Stop()
	close(signalChan)

	return bot.Deals(), bot.BalanceHistory(), nil
}

// DCA - Dollar Cost Averaging
func (t *TradingBotService) BacktestDCA(
	md domain.MarketData,
	addition int,
	commissionPercent float64,
	slippagePercent float64,
	from time.Time,
	to time.Time,
) (deals []domain.Deal, baseInvestments float64, resultBalance float64, err error) {
	candles, err := t.mdService.GetCandlesByTime(md, from, to)
	if err != nil {
		return nil, 0, 0, err
	}

	if len(candles) < 2 {
		return nil, 0, 0, fmt.Errorf("not enough data")
	}

	currBalance := 0.0
	currCount := 0

	for _, candle := range candles {
		log.Println(candle.Open, candle.OpenTime)
		baseInvestments += float64(addition)
		currBalance += float64(addition)
		count := int(currBalance / candle.Open)
		currCount += count
		currBalance = math.Mod(currBalance, candle.Open)

		basePrice := float64(count) * candle.Open
		basePrice *= 1 + slippagePercent

		commision := basePrice * commissionPercent

		deal := domain.Deal{
			Direction:  domain.DealDirection_BUY,
			Time:       candle.OpenTime,
			Price:      candle.Open,
			Count:      count,
			LotPrice:   basePrice + commision,
			Commission: commision,
		}
		deals = append(deals, deal)
	}

	lastCandle := candles[len(candles)-1]

	basePrice := float64(currCount) * lastCandle.Close
	basePrice *= 1 - slippagePercent

	commision := basePrice * commissionPercent

	deal := domain.Deal{
		Direction:  domain.DealDirection_SELL,
		Time:       lastCandle.CloseTime,
		Price:      lastCandle.Close,
		Count:      currCount,
		LotPrice:   basePrice + commision,
		Commission: commision,
	}
	deals = append(deals, deal)

	currBalance += basePrice - commision

	return deals, baseInvestments, currBalance, nil
}
