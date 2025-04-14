package marketdata

type SubscribeInfo struct {
	intrumentID string
	interval SubscriptionInterval
}

type Candle struct {
	High float64
	Low  float64
}

type SubscriptionInterval int32

const (
	ONE_MINUTE      SubscriptionInterval = 0  //Минутные свечи.
	FIVE_MINUTES    SubscriptionInterval = 1  //Пятиминутные свечи.
	FIFTEEN_MINUTES SubscriptionInterval = 2  //Пятнадцатиминутные свечи.
	ONE_HOUR        SubscriptionInterval = 3  //Часовые свечи.
	ONE_DAY         SubscriptionInterval = 4  //Дневные свечи.
	TWO_MIN           SubscriptionInterval = 5  //Двухминутные свечи.
	THREE_MIN           SubscriptionInterval = 6  //Трехминутные свечи.
	TEN_MIN          SubscriptionInterval = 7  //Десятиминутные свечи.
	THERTY_MIN          SubscriptionInterval = 8  //Тридцатиминутные свечи.
	TWO_HOUR          SubscriptionInterval = 9 //Двухчасовые свечи.
	FOUR_HOUR          SubscriptionInterval = 10 //Четырехчасовые свечи.
	WEEK            SubscriptionInterval = 11 //Недельные свечи.
	MONTH           SubscriptionInterval = 12 //Месячные свечи.
)