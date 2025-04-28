package marketdata

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Reensef/sigmasage/pkg/utils"

	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// TinkoffMarketdata implements an observer that distributes candle data to subscribers
type TinkoffMarketdata struct {
	token                         string
	mu                            sync.RWMutex
	mdStream                      *investgo.MarketDataStream
	mdService                     *investgo.MarketDataServiceClient
	candleSubscribers             map[MarketDataInfo][]chan Candle
	candlesNotifyCancel           context.CancelFunc
	isNotifyingCandlesSubscribers bool
}

func NewTinkoffMarketdata(token string) (*TinkoffMarketdata, error) {
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

	mdStreamClient := client.NewMarketDataStreamClient()

	mdStream, err := mdStreamClient.MarketDataStream()
	if err != nil {
		return nil, err
	}

	mdService := client.NewMarketDataServiceClient()

	return &TinkoffMarketdata{
		token:             token,
		candleSubscribers: make(map[MarketDataInfo][]chan Candle),
		mdStream:          mdStream,
		mdService:         mdService,
	}, nil
}

// Subscribe returns a channel that will receive candle data for the specified instrument
func (t *TinkoffMarketdata) SubscribeCandles(marketDataInfo MarketDataInfo) (<-chan Candle, error) {
	// t.mu.Lock()
	// defer t.mu.Unlock()

	ch := make(chan Candle, 100)

	if _, exists := t.candleSubscribers[marketDataInfo]; !exists {
		t.candleSubscribers[marketDataInfo] = make([]chan Candle, 0)
	}
	t.candleSubscribers[marketDataInfo] = append(t.candleSubscribers[marketDataInfo], ch)

	candlesChan, err := t.mdStream.SubscribeCandle(
		[]string{marketDataInfo.IntrumentID},
		t.convertToSubscriptionInterval(marketDataInfo.Interval),
		true,
		nil,
	)

	if err != nil {
		return nil, err
	}

	if !t.isNotifyingCandlesSubscribers {
		t.startNotifyingCandlesSubscribers(candlesChan)
		t.isNotifyingCandlesSubscribers = true
	}

	return ch, nil
}

// Unsubscribe removes a channel from the subscribers list
func (t *TinkoffMarketdata) UnsubscribeCandles(marketDataInfo MarketDataInfo, ch <-chan Candle) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	ok := false

	if subscribers, exists := t.candleSubscribers[marketDataInfo]; exists {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				t.candleSubscribers[marketDataInfo] = append(subscribers[:i], subscribers[i+1:]...)
				close(subscriber)
				ok = true
				break
			}
		}

		if len(t.candleSubscribers[marketDataInfo]) == 0 {
			delete(t.candleSubscribers, marketDataInfo)
		}
	}

	return ok
}

func (t *TinkoffMarketdata) startNotifyingCandlesSubscribers(candlesChan <-chan *pb.Candle) {
	var ctx context.Context
	ctx, t.candlesNotifyCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		// defer wg.Done()
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
			case candle, ok := <-candlesChan:
				if !ok {
					return
				}
				fmt.Println("Candle high price = ", candle.GetHigh().ToFloat())
			}
		}
	}(ctx)
}

func (t *TinkoffMarketdata) GetCandles(
	marketDataInfo MarketDataInfo,
	from time.Time,
	to time.Time,
) ([]Candle, error) {
	result := make([]Candle, 0)

	resp, err := t.mdService.GetCandles(
		marketDataInfo.IntrumentID,
		t.convertToCandleInterval(marketDataInfo.Interval),
		from,
		to,
		pb.GetCandlesRequest_CANDLE_SOURCE_EXCHANGE,
		0,
	)

	if err != nil {
		return nil, err
	}

	for _, candle := range resp.GetCandles() {
		result = append(result, Candle{
			Time:   candle.GetTime().AsTime(),
			Open:   candle.GetOpen().ToFloat(),
			High:   candle.GetHigh().ToFloat(),
			Low:    candle.GetLow().ToFloat(),
			Close:  candle.GetClose().ToFloat(),
			Volume: float64(candle.GetVolume()),
		})
	}

	return result, nil
}

func (t *TinkoffMarketdata) convertToSubscriptionInterval(interval MarketDataInterval) pb.SubscriptionInterval {
	switch interval {
	case UNSPECIFIED:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_UNSPECIFIED
	case ONE_MINUTE:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE
	case TWO_MIN:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_2_MIN
	case THREE_MIN:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_3_MIN
	case FIVE_MINUTES:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_FIVE_MINUTES
	case TEN_MIN:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_10_MIN
	case FIFTEEN_MINUTES:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_FIFTEEN_MINUTES
	case THERTY_MIN:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_30_MIN
	case ONE_HOUR:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_HOUR
	case TWO_HOUR:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_2_HOUR
	case FOUR_HOUR:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_4_HOUR
	case ONE_DAY:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_DAY
	case WEEK:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_WEEK
	case MONTH:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_MONTH
	default:
		return pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_UNSPECIFIED
	}
}

func (t *TinkoffMarketdata) convertToCandleInterval(interval MarketDataInterval) pb.CandleInterval {
	switch interval {
	case UNSPECIFIED:
		return pb.CandleInterval_CANDLE_INTERVAL_UNSPECIFIED
	case ONE_MINUTE:
		return pb.CandleInterval_CANDLE_INTERVAL_1_MIN
	case TWO_MIN:
		return pb.CandleInterval_CANDLE_INTERVAL_2_MIN
	case THREE_MIN:
		return pb.CandleInterval_CANDLE_INTERVAL_3_MIN
	case FIVE_MINUTES:
		return pb.CandleInterval_CANDLE_INTERVAL_5_MIN
	case TEN_MIN:
		return pb.CandleInterval_CANDLE_INTERVAL_10_MIN
	case FIFTEEN_MINUTES:
		return pb.CandleInterval_CANDLE_INTERVAL_15_MIN
	case THERTY_MIN:
		return pb.CandleInterval_CANDLE_INTERVAL_30_MIN
	case ONE_HOUR:
		return pb.CandleInterval_CANDLE_INTERVAL_HOUR
	case TWO_HOUR:
		return pb.CandleInterval_CANDLE_INTERVAL_2_HOUR
	case FOUR_HOUR:
		return pb.CandleInterval_CANDLE_INTERVAL_4_HOUR
	case ONE_DAY:
		return pb.CandleInterval_CANDLE_INTERVAL_DAY
	case WEEK:
		return pb.CandleInterval_CANDLE_INTERVAL_WEEK
	case MONTH:
		return pb.CandleInterval_CANDLE_INTERVAL_MONTH
	default:
		return pb.CandleInterval_CANDLE_INTERVAL_UNSPECIFIED
	}
}
