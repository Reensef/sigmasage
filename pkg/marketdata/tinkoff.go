package marketdata

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Reensef/sigmasage/pkg/utils"

	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// TinkoffMarketdata implements an observer that distributes candle data to subscribers
type TinkoffMarketdata struct {
	token               string
	mu                  sync.RWMutex
	mdStream            *investgo.MarketDataStream
	candleSubscribers   map[SubscribeInfo][]chan Candle
	candlesChanCancel   context.CancelFunc
	areCandlesStreaming bool
}

func New(token string) *TinkoffMarketdata {
	logger := utils.TinkoffLogger{}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	config := investgo.Config{
		EndPoint:        "invest-public-api.tinkoff.ru:443",
		Token:           "t.b8rY7Kv7HppxKdwCynTNbZbJ6kCzR09EbEQf5gAsJhGIOVU1oteZAit8eRdInD1pFxC8h8PvnAtumjuvcDn3pg",
		MaxRetries:      3,
		AppName:         "sigmasage",
		DisableAllRetry: false,
	}

	client, err := investgo.NewClient(ctx, config, logger)
	if err != nil {
		panic(err)
	}

	mdStreamClient := client.NewMarketDataStreamClient()
	mdStream, err := mdStreamClient.MarketDataStream()
	if err != nil {
		panic(err)
	}

	return &TinkoffMarketdata{
		token:             token,
		candleSubscribers: make(map[SubscribeInfo][]chan Candle),
		mdStream:          mdStream,
	}
}

// Subscribe returns a channel that will receive candle data for the specified instrument
func (t *TinkoffMarketdata) SubscribeCandles(subscribeInfo SubscribeInfo) (<-chan Candle, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	ch := make(chan Candle, 100)

	if _, exists := t.candleSubscribers[subscribeInfo]; !exists {
		t.candleSubscribers[subscribeInfo] = make([]chan Candle, 0)
	}
	t.candleSubscribers[subscribeInfo] = append(t.candleSubscribers[subscribeInfo], ch)

	t.startStreamCandles(subscribeInfo)

	return ch, nil
}

// Unsubscribe removes a channel from the subscribers list
func (t *TinkoffMarketdata) UnsubscribeCandles(subscribeInfo SubscribeInfo, ch <-chan Candle) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if subscribers, exists := t.candleSubscribers[subscribeInfo]; exists {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				t.candleSubscribers[subscribeInfo] = append(subscribers[:i], subscribers[i+1:]...)
				close(subscriber)
				break
			}
		}

		if len(t.candleSubscribers[subscribeInfo]) == 0 {
			delete(t.candleSubscribers, subscribeInfo)
		}
	}
}

// streamCandles starts a gRPC stream for the specified instrument and interval
func (t *TinkoffMarketdata) startStreamCandles(subscribeInfo SubscribeInfo) {
	tinkoffInterval := t.convertToTinkoffCandleInterval(subscribeInfo.interval)

	// Start streaming
	candlesChan, err := t.mdStream.SubscribeCandle(
		[]string{subscribeInfo.intrumentID},
		tinkoffInterval,
		true,
		nil,
	)
	if err != nil {
		// TODO: Handle error properly
		return
	}

	if t.areCandlesStreaming {
		return
	}

	t.areCandlesStreaming = true

	var ctx context.Context
	ctx, t.candlesChanCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

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

func (t *TinkoffMarketdata) convertToTinkoffCandleInterval(interval SubscriptionInterval) pb.SubscriptionInterval {
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
