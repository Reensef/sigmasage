package marketdata

type SubscribeInfo struct {
	intrumentID string
	interval    SubscriptionInterval
}

type Candle struct {
	SubscribeInfo SubscribeInfo
	High          float64
	Low           float64
}

type SubscriptionInterval int32

const (
	UNSPECIFIED     SubscriptionInterval = iota
	ONE_MINUTE      SubscriptionInterval = iota
	TWO_MIN         SubscriptionInterval = iota
	THREE_MIN       SubscriptionInterval = iota
	FIVE_MINUTES    SubscriptionInterval = iota
	TEN_MIN         SubscriptionInterval = iota
	FIFTEEN_MINUTES SubscriptionInterval = iota
	THERTY_MIN      SubscriptionInterval = iota
	ONE_HOUR        SubscriptionInterval = iota
	TWO_HOUR        SubscriptionInterval = iota
	FOUR_HOUR       SubscriptionInterval = iota
	ONE_DAY         SubscriptionInterval = iota
	WEEK            SubscriptionInterval = iota
	MONTH           SubscriptionInterval = iota
)
