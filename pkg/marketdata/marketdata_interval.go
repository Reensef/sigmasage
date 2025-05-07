package marketdata

import "time"

type MarketDataInterval int32

const (
	UNSPECIFIED     MarketDataInterval = iota
	ONE_MINUTE      MarketDataInterval = iota
	TWO_MIN         MarketDataInterval = iota
	THREE_MIN       MarketDataInterval = iota
	FIVE_MINUTES    MarketDataInterval = iota
	TEN_MIN         MarketDataInterval = iota
	FIFTEEN_MINUTES MarketDataInterval = iota
	THERTY_MIN      MarketDataInterval = iota
	ONE_HOUR        MarketDataInterval = iota
	TWO_HOUR        MarketDataInterval = iota
	FOUR_HOUR       MarketDataInterval = iota
	ONE_DAY         MarketDataInterval = iota
	WEEK            MarketDataInterval = iota
	MONTH           MarketDataInterval = iota
)

func ConvertMarketDataIntervalToTime(interval MarketDataInterval) time.Duration {
	switch interval {
	case ONE_MINUTE:
		return time.Minute
	case TWO_MIN:
		return 2 * time.Minute
	case THREE_MIN:
		return 3 * time.Minute
	case FIVE_MINUTES:
		return 5 * time.Minute
	case TEN_MIN:
		return 10 * time.Minute
	case FIFTEEN_MINUTES:
		return 15 * time.Minute
	case THERTY_MIN:
		return 30 * time.Minute
	case ONE_HOUR:
		return time.Hour
	case TWO_HOUR:
		return 2 * time.Hour
	case FOUR_HOUR:
		return 4 * time.Hour
	case ONE_DAY:
		return 24 * time.Hour
	case WEEK:
		return 7 * 24 * time.Hour
	case MONTH:
		return 30 * 24 * time.Hour
	default:
		return 0
	}
}

// TODO Тесты!!!
// Считает длительность в зависимости от интервала и количества, учитывая только рабочие часы
func AdjustDurationForWorkingHours(interval MarketDataInterval, count int) time.Duration {
	normalize := func(duration time.Duration) time.Duration {
		daysCount := int(duration.Hours()) / 8

		if daysCount > 7 {
			duration += time.Duration((daysCount/7)*2*24) * time.Hour
		}

		return duration
	}

	switch interval {
	case ONE_MINUTE:
		return normalize(time.Minute * time.Duration(count))
	case TWO_MIN:
		return normalize(2 * time.Minute * time.Duration(count))
	case THREE_MIN:
		return normalize(3 * time.Minute * time.Duration(count))
	case FIVE_MINUTES:
		return normalize(5 * time.Minute * time.Duration(count))
	case TEN_MIN:
		return normalize(10 * time.Minute * time.Duration(count))
	case FIFTEEN_MINUTES:
		return normalize(15 * time.Minute * time.Duration(count))
	case THERTY_MIN:
		return normalize(30 * time.Minute * time.Duration(count))
	case ONE_HOUR:
		return normalize(time.Hour * time.Duration(count))
	case TWO_HOUR:
		return normalize(2 * time.Hour * time.Duration(count))
	case FOUR_HOUR:
		return normalize(4 * time.Hour * time.Duration(count))
	case ONE_DAY:
		return normalize(24 * time.Hour * time.Duration(count))
	case WEEK:
		return normalize(7 * 24 * time.Hour * time.Duration(count))
	case MONTH:
		return normalize(30 * 24 * time.Hour * time.Duration(count))
	default:
		return 0
	}
}
