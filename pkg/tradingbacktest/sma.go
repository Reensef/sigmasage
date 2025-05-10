package tradingbacktest

import "github.com/Reensef/sigmasage/pkg/strategy"

// Returns balance after all signals are processed
// Commission is in percents
func SMACBacktest(signals []strategy.SMACSignal, balance float64, commission float64) float64 {
	count := 0

	for _, signal := range signals {
		marketData := signal.Info.MarketData
		price := marketData.GetPrice(signal.Time)

		if (signal.SignalType == strategy.SMACSignalType_UPWARDS_CROSS) {
			balance -= 
		} else if signal.SignalType == strategy.SMACSignalType_DOWNWARDS_CROSS {
			balance = balance * (1 + commission/100)
		}
	}
	return balance
}
