package marketdata

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Reensef/sigmasage/pkg/domain"
	"github.com/Reensef/sigmasage/pkg/utils"

	"slices"

	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// TinkoffMarketDataProvider implements an observer that distributes candle data to subscribers
type TinkoffMarketDataProvider struct {
	token                         string
	mu                            sync.RWMutex
	mdStream                      *investgo.MarketDataStream
	mdService                     *investgo.MarketDataServiceClient
	instrumentService             *investgo.InstrumentsServiceClient
	candleSubscribers             map[domain.MarketData][]chan domain.Candle
	candlesNotifyCancel           context.CancelFunc
	isNotifyingCandlesSubscribers bool
}

func NewTinkoffMarketDataProvider(token string) (*TinkoffMarketDataProvider, error) {
	logger := utils.TinkoffLogger{}

	config := investgo.Config{
		EndPoint:        "invest-public-api.tinkoff.ru:443",
		Token:           token,
		MaxRetries:      3,
		AppName:         "sigmasage",
		DisableAllRetry: false,
	}

	client, err := investgo.NewClient(context.Background(), config, logger)
	if err != nil {
		return nil, err
	}

	instrumentsService := client.NewInstrumentsServiceClient()

	instrResp, err := instrumentsService.ShareByTicker("YDEX", "TQBR")
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		ins := instrResp.GetInstrument()
		fmt.Printf("Инструменты по запросу YDEX - %v %v\n", ins.GetName(), ins.GetUid())
	}

	instrumentsService = client.NewInstrumentsServiceClient()

	instrResp, err = instrumentsService.ShareByTicker("LKOH", "TQBR")
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		ins := instrResp.GetInstrument()
		fmt.Printf("Инструменты по запросу LKOH - %v %v\n", ins.GetName(), ins.GetUid())
	}

	mdStreamClient := client.NewMarketDataStreamClient()

	mdStream, err := mdStreamClient.MarketDataStream()
	if err != nil {
		return nil, err
	}

	mdService := client.NewMarketDataServiceClient()

	return &TinkoffMarketDataProvider{
		token:             token,
		candleSubscribers: make(map[domain.MarketData][]chan domain.Candle),
		mdStream:          mdStream,
		mdService:         mdService,
		instrumentService: instrumentsService,
	}, nil
}

func (t *TinkoffMarketDataProvider) SubscribeCandles(
	marketDataInfo domain.MarketData,
) (<-chan domain.Candle, error) {
	ch := make(chan domain.Candle, 100)

	if _, exists := t.candleSubscribers[marketDataInfo]; !exists {
		t.candleSubscribers[marketDataInfo] =
			make([]chan domain.Candle, 0)
	}
	t.candleSubscribers[marketDataInfo] =
		append(t.candleSubscribers[marketDataInfo], ch)

	candlesChan, err := t.mdStream.SubscribeCandle(
		[]string{marketDataInfo.ID},
		t.convertToSubscriptionInterval(marketDataInfo.Interval),
		true,
		nil,
	)

	if err != nil {
		return nil, err
	}

	if !t.isNotifyingCandlesSubscribers {
		t.startNotifyingCandlesSubscribers(
			marketDataInfo,
			candlesChan,
		)
		t.isNotifyingCandlesSubscribers = true
	}

	return ch, nil
}

func (t *TinkoffMarketDataProvider) UnsubscribeCandles(
	marketDataInfo domain.MarketData,
	ch <-chan domain.Candle,
) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if subscribers, exists :=
		t.candleSubscribers[marketDataInfo]; exists {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				t.candleSubscribers[marketDataInfo] =
					slices.Delete(subscribers, i, i+1)
				close(subscriber)

				l := len(t.candleSubscribers[marketDataInfo])
				if l == 0 {
					delete(t.candleSubscribers, marketDataInfo)
				}

				return nil
			}
		}
	}

	return fmt.Errorf("undefined subscriber")
}

func (t *TinkoffMarketDataProvider) GetCandlesByTime(
	marketData domain.MarketData,
	from time.Time,
	to time.Time,
) ([]domain.Candle, error) {
	result := make([]domain.Candle, 0)

	resp, err := t.mdService.GetHistoricCandles(
		&investgo.GetHistoricCandlesRequest{
			Instrument: marketData.ID,
			Interval: t.convertToCandleInterval(
				marketData.Interval,
			),
			From:   from,
			To:     to,
			Source: pb.GetCandlesRequest_CANDLE_SOURCE_INCLUDE_WEEKEND,
		},
	)

	if err != nil {
		return nil, err
	}

	for _, candle := range resp {
		result = append(result, domain.Candle{
			MarketData: marketData,
			OpenTime:   candle.GetTime().AsTime(),
			CloseTime: candle.GetTime().AsTime().Add(
				time.Duration(
					ConvertMarketDataIntervalToTime(
						marketData.Interval,
					),
				),
			),
			Open:   candle.GetOpen().ToFloat(),
			High:   candle.GetHigh().ToFloat(),
			Low:    candle.GetLow().ToFloat(),
			Close:  candle.GetClose().ToFloat(),
			Volume: float64(candle.GetVolume()),
		})
	}

	return result, nil
}

func (t *TinkoffMarketDataProvider) GetCandlesByCount(
	marketData domain.MarketData,
	last time.Time,
	count int,
) ([]domain.Candle, error) {
	result := make([]domain.Candle, 0)

	first := last.Add(-time.Hour * 24 * 14) // 2 weeks ago
	first = first.Add(
		-AdjustDurationForWorkingHours(
			marketData.Interval, count,
		),
	) // only working hours

	resp, err := t.mdService.GetHistoricCandles(
		&investgo.GetHistoricCandlesRequest{
			Instrument: marketData.ID,
			Interval: t.convertToCandleInterval(
				marketData.Interval,
			),
			From:   first,
			To:     last,
			Source: pb.GetCandlesRequest_CANDLE_SOURCE_INCLUDE_WEEKEND,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(resp) < count {
		return nil, fmt.Errorf(
			"error getting history data by count",
		)
	}

	for _, candle := range resp[len(resp)-count:] {
		result = append(result, domain.Candle{
			MarketData: marketData,
			OpenTime:   candle.GetTime().AsTime(),
			CloseTime: candle.GetTime().AsTime().Add(
				time.Duration(
					ConvertMarketDataIntervalToTime(
						marketData.Interval,
					),
				),
			),
			Open:   candle.GetOpen().ToFloat(),
			High:   candle.GetHigh().ToFloat(),
			Low:    candle.GetLow().ToFloat(),
			Close:  candle.GetClose().ToFloat(),
			Volume: float64(candle.GetVolume()),
		})
	}

	return result, nil
}

func (t *TinkoffMarketDataProvider) startNotifyingCandlesSubscribers(marketdata domain.MarketData, candlesChan <-chan *pb.Candle) {
	var ctx context.Context
	ctx, t.candlesNotifyCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		err := t.mdStream.Listen()
		if err != nil {
			log.Printf("Error listening to candles: %v", err)
		}
	}()

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case pbCandle, ok := <-candlesChan:
				if !ok {
					return
				}

				candle := domain.Candle{
					MarketData: marketdata,
					OpenTime:   pbCandle.GetTime().AsTime(),
					CloseTime:  pbCandle.GetTime().AsTime().Add(time.Duration(ConvertMarketDataIntervalToTime(marketdata.Interval))),
					Open:       pbCandle.GetOpen().ToFloat(),
					High:       pbCandle.GetHigh().ToFloat(),
					Low:        pbCandle.GetLow().ToFloat(),
					Close:      pbCandle.GetClose().ToFloat(),
					Volume:     float64(pbCandle.GetVolume()),
				}
				t.mu.RLock()
				for _, subscriber := range t.candleSubscribers[marketdata] {
					subscriber <- candle
				}
				t.mu.RUnlock()
			}
		}
	}(ctx)
}

func (t *TinkoffMarketDataProvider) convertToSubscriptionInterval(interval domain.MarketDataInterval) pb.SubscriptionInterval {
	switch interval {
	case domain.MarketDataInterval_UNSPECIFIED:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_UNSPECIFIED
	case domain.MarketDataInterval_ONE_MINUTE:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE
	case domain.MarketDataInterval_TWO_MIN:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_2_MIN
	case domain.MarketDataInterval_THREE_MIN:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_3_MIN
	case domain.MarketDataInterval_FIVE_MINUTES:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_FIVE_MINUTES
	case domain.MarketDataInterval_TEN_MIN:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_10_MIN
	case domain.MarketDataInterval_FIFTEEN_MINUTES:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_FIFTEEN_MINUTES
	case domain.MarketDataInterval_THERTY_MIN:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_30_MIN
	case domain.MarketDataInterval_ONE_HOUR:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_HOUR
	case domain.MarketDataInterval_TWO_HOUR:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_2_HOUR
	case domain.MarketDataInterval_FOUR_HOUR:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_4_HOUR
	case domain.MarketDataInterval_ONE_DAY:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_DAY
	case domain.MarketDataInterval_WEEK:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_WEEK
	case domain.MarketDataInterval_MONTH:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_MONTH
	default:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_UNSPECIFIED
	}
}

func (t *TinkoffMarketDataProvider) convertToCandleInterval(interval domain.MarketDataInterval) pb.CandleInterval {
	switch interval {
	case domain.MarketDataInterval_UNSPECIFIED:
		return pb.CandleInterval_CANDLE_INTERVAL_UNSPECIFIED
	case domain.MarketDataInterval_ONE_MINUTE:
		return pb.CandleInterval_CANDLE_INTERVAL_1_MIN
	case domain.MarketDataInterval_TWO_MIN:
		return pb.CandleInterval_CANDLE_INTERVAL_2_MIN
	case domain.MarketDataInterval_THREE_MIN:
		return pb.CandleInterval_CANDLE_INTERVAL_3_MIN
	case domain.MarketDataInterval_FIVE_MINUTES:
		return pb.CandleInterval_CANDLE_INTERVAL_5_MIN
	case domain.MarketDataInterval_TEN_MIN:
		return pb.CandleInterval_CANDLE_INTERVAL_10_MIN
	case domain.MarketDataInterval_FIFTEEN_MINUTES:
		return pb.CandleInterval_CANDLE_INTERVAL_15_MIN
	case domain.MarketDataInterval_THERTY_MIN:
		return pb.CandleInterval_CANDLE_INTERVAL_30_MIN
	case domain.MarketDataInterval_ONE_HOUR:
		return pb.CandleInterval_CANDLE_INTERVAL_HOUR
	case domain.MarketDataInterval_TWO_HOUR:
		return pb.CandleInterval_CANDLE_INTERVAL_2_HOUR
	case domain.MarketDataInterval_FOUR_HOUR:
		return pb.CandleInterval_CANDLE_INTERVAL_4_HOUR
	case domain.MarketDataInterval_ONE_DAY:
		return pb.CandleInterval_CANDLE_INTERVAL_DAY
	case domain.MarketDataInterval_WEEK:
		return pb.CandleInterval_CANDLE_INTERVAL_WEEK
	case domain.MarketDataInterval_MONTH:
		return pb.CandleInterval_CANDLE_INTERVAL_MONTH
	default:
		return pb.CandleInterval_CANDLE_INTERVAL_UNSPECIFIED
	}
}
