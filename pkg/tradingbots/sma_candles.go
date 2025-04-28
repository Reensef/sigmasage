package tradingbots

import (
	"log"
	"time"

	"github.com/Reensef/sigmasage/pkg/exchange"
	"github.com/Reensef/sigmasage/pkg/indicators"
	"github.com/Reensef/sigmasage/pkg/marketdata"
)

type SMACandlesBot struct {
	candlesData  marketdata.CandleProvider
	exchanger    exchange.Exchanger
	instrumentID string
	interval     marketdata.MarketDataInterval
	length       int
	sum          float64
	boughtCount  int
	bought       bool
	running      bool
}

func NewSMACandlesBot(
	candlesData marketdata.CandleProvider,
	exchanger exchange.Exchanger,
	instrumentID string,
	interval marketdata.MarketDataInterval,
	length int,
	sum float64,
) *SMACandlesBot {
	return &SMACandlesBot{
		candlesData:  candlesData,
		exchanger:    exchanger,
		instrumentID: instrumentID,
		interval:     interval,
		length:       length,
		sum:          sum,
	}
}

func (bot *SMACandlesBot) RunReal() {
	if bot.running {
		return
	}
	bot.running = true

	sma := indicators.NewSMA(bot.length)
	wasHistoryLoaded := false

	go func() {
		historicalCandles, err := bot.candlesData.GetCandles(marketdata.MarketDataInfo{
			IntrumentID: bot.instrumentID,
			Interval:    bot.interval,
		}, time.Now().Add(-time.Duration(bot.length)*marketdata.ConvertMarketDataIntervalToTime(bot.interval)), time.Now())
		if err != nil {
			log.Println("Error getting historical candles:", err)
			return
		}

		for _, candle := range historicalCandles {
			sma.Update(candle.Close)
		}
		wasHistoryLoaded = true
	}()

	candlesChan, err := bot.candlesData.SubscribeCandles(marketdata.MarketDataInfo{
		IntrumentID: bot.instrumentID,
		Interval:    bot.interval,
	})
	if err != nil {
		log.Println("Error subscribing to candles:", err)
		return
	}

	diff := make([]marketdata.Candle, 0, bot.length)

	for {
		candle := <-candlesChan
		diff = append(diff, candle)

		if wasHistoryLoaded {
			break
		}
	}

	// Догружаем данные
	for _, candle := range diff {
		sma.Update(candle.Close)
	}

	bot.run(candlesChan, sma)
}

func (bot *SMACandlesBot) RunHistory(from time.Time, to time.Time) {
	if bot.running {
		return
	}
	bot.running = true

	sma := indicators.NewSMA(bot.length)

	mdInfo := marketdata.MarketDataInfo{
		IntrumentID: bot.instrumentID,
		Interval:    bot.interval,
	}

	historicalCandles, err := bot.candlesData.GetCandles(
		mdInfo,
		from.Add(-time.Duration(bot.length)*marketdata.ConvertMarketDataIntervalToTime(bot.interval)),
		to,
	)
	if err != nil {
		log.Println("Error getting historical candles:", err)
		return
	}

	for _, candle := range historicalCandles[:bot.length] {
		sma.Update(candle.Close)
	}

	candlesChan := make(chan marketdata.Candle, 100)

	go func() {
		for _, candle := range historicalCandles[bot.length:] {
			candlesChan <- candle
		}
	}()

	bot.run(candlesChan, sma)
}

// Нужно добавить буфер для ордеров
func (bot *SMACandlesBot) run(candlesChan <-chan marketdata.Candle, sma *indicators.SMA) {
	for {
		candle := <-candlesChan
		smaValue := sma.Update(candle.Close)

		if candle.Close > smaValue {
			if !bot.bought {
				count := int(bot.sum / candle.Low)
				orderInfo := exchange.OrderInfo{
					InstrumentID: bot.instrumentID,
					Count:        count,
					Price:        candle.Low,
				}

				turnover, err := bot.exchanger.Buy(orderInfo)
				if err != nil {
					log.Println("Error buying:", err)
					continue
				}
				log.Printf("Buy Candle: SMA %v, Order info %v", smaValue, orderInfo)
				log.Printf("Sum: %v", bot.sum)
				bot.sum -= turnover
				bot.boughtCount += count
				bot.bought = true
			}
		} else if candle.Close < smaValue {
			if bot.bought {
				orderInfo := exchange.OrderInfo{
					InstrumentID: bot.instrumentID,
					Count:        bot.boughtCount,
					Price:        candle.Low,
				}

				turnover, err := bot.exchanger.Sell(orderInfo)
				if err != nil {
					log.Println("Error selling:", err)
					continue
				}
				log.Printf("Sell Candle: SMA %v, Order info %v", smaValue, orderInfo)
				log.Printf("Sum: %v", bot.sum)
				bot.sum += turnover
				bot.boughtCount = 0
				bot.bought = false
			}
		}
	}
}
