package marketdata

import (
	"time"

	"github.com/Reensef/sigmasage/pkg/domain"
)

func ConvertMarketDataIntervalToTime(interval domain.MarketDataInterval) time.Duration {
	switch interval {
	case domain.MarketDataInterval_ONE_MINUTE:
		return time.Minute
	case domain.MarketDataInterval_TWO_MIN:
		return 2 * time.Minute
	case domain.MarketDataInterval_THREE_MIN:
		return 3 * time.Minute
	case domain.MarketDataInterval_FIVE_MINUTES:
		return 5 * time.Minute
	case domain.MarketDataInterval_TEN_MIN:
		return 10 * time.Minute
	case domain.MarketDataInterval_FIFTEEN_MINUTES:
		return 15 * time.Minute
	case domain.MarketDataInterval_THERTY_MIN:
		return 30 * time.Minute
	case domain.MarketDataInterval_ONE_HOUR:
		return time.Hour
	case domain.MarketDataInterval_TWO_HOUR:
		return 2 * time.Hour
	case domain.MarketDataInterval_FOUR_HOUR:
		return 4 * time.Hour
	case domain.MarketDataInterval_ONE_DAY:
		return 24 * time.Hour
	case domain.MarketDataInterval_WEEK:
		return 7 * 24 * time.Hour
	case domain.MarketDataInterval_MONTH:
		return 30 * 24 * time.Hour
	default:
		return 0
	}
}

// TODO Тесты!!!
// Считает длительность в зависимости от интервала и количества, учитывая только рабочие часы
func AdjustDurationForWorkingHours(interval domain.MarketDataInterval, count int) time.Duration {
	normalize := func(duration time.Duration) time.Duration {
		daysCount := int(duration.Hours()) / 8

		if daysCount > 7 {
			duration += time.Duration((daysCount/7)*2*24) * time.Hour
		}

		return duration
	}

	switch interval {
	case domain.MarketDataInterval_ONE_MINUTE:
		return normalize(time.Minute * time.Duration(count))
	case domain.MarketDataInterval_TWO_MIN:
		return normalize(2 * time.Minute * time.Duration(count))
	case domain.MarketDataInterval_THREE_MIN:
		return normalize(3 * time.Minute * time.Duration(count))
	case domain.MarketDataInterval_FIVE_MINUTES:
		return normalize(5 * time.Minute * time.Duration(count))
	case domain.MarketDataInterval_TEN_MIN:
		return normalize(10 * time.Minute * time.Duration(count))
	case domain.MarketDataInterval_FIFTEEN_MINUTES:
		return normalize(15 * time.Minute * time.Duration(count))
	case domain.MarketDataInterval_THERTY_MIN:
		return normalize(30 * time.Minute * time.Duration(count))
	case domain.MarketDataInterval_ONE_HOUR:
		return normalize(time.Hour * time.Duration(count))
	case domain.MarketDataInterval_TWO_HOUR:
		return normalize(2 * time.Hour * time.Duration(count))
	case domain.MarketDataInterval_FOUR_HOUR:
		return normalize(4 * time.Hour * time.Duration(count))
	case domain.MarketDataInterval_ONE_DAY:
		return normalize(24 * time.Hour * time.Duration(count))
	case domain.MarketDataInterval_WEEK:
		return normalize(7 * 24 * time.Hour * time.Duration(count))
	case domain.MarketDataInterval_MONTH:
		return normalize(30 * 24 * time.Hour * time.Duration(count))
	default:
		return 0
	}
}
