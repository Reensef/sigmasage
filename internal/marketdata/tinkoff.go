package marketdata

// Реализует обсервер, рассылает всем подписчикам
// информацию о свечах по инструменту с заданным интервалом
// TODO нужно как то придумать как получать внутри пачку и распределять её между подписотой
type TinkoffCandleMarketdata struct {
	userID        int64
	token         string
	instrumentsID []string
	interval      SubscriptionInterval
}

func New(token string) *TinkoffCandleMarketdata {
	return &TinkoffCandleMarketdata{token: token}
}

func (t *TinkoffCandleMarketdata) Subscribe() (<-chan Candle, error) {

}
