package marketdata

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Reensef/sigmasage/pkg/utils"

	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// TinkoffMarketdata implements an observer that distributes candle data to subscribers
type TinkoffMarketdata struct {
	token             string
	candleSubscribers map[SubscribeInfo][]chan Candle // map of instrumentID to channel of candles
	mu                sync.RWMutex
	mdStreamClient    *investgo.MarketDataStreamClient
}

func New(token string) *TinkoffMarketdata {
	logger := utils.TinkoffLogger{}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	config := investgo.Config{
		EndPoint:        "invest-public-api.tinkoff.ru:443",
		Token:           token,
		MaxRetries:      3,
		AppName:         "sigmasage",
		DisableAllRetry: false,
	}

	client, err := investgo.NewClient(ctx, config, logger)
	if err != nil {
		panic(err)
	}

	mdStreamClient := client.NewMarketDataStreamClient()

	return &TinkoffMarketdata{
		token:             token,
		candleSubscribers: make(map[SubscribeInfo][]chan Candle),
		mdStreamClient:    mdStreamClient,
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

	if len(t.candleSubscribers[subscribeInfo]) == 1 {
		go t.streamCandles(subscribeInfo)
	}

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
func (t *TinkoffMarketdata) streamCandles(subscribeInfo SubscribeInfo) {
	ctx := context.Background()

	tinkoffInterval := t.convertToTinkoffCandleInterval(subscribeInfo.interval)

	// Create subscription request
	req := &investgo.CandleInstrument{
		Figi:     instrumentID,
		Interval: tinkoffInterval,
	}

	// Start streaming
	stream, err := t.marketDataClient.GetCandles(ctx, req)
	if err != nil {
		// TODO: Handle error properly
		return
	}

	for {
		candle, err := stream.Recv()
		if err != nil {
			// TODO: Handle error properly
			return
		}

		// Convert Tinkoff candle to our Candle type
		ourCandle := Candle{
			InstrumentID: instrumentID,
			Interval:     interval,
			High:         candle.GetHigh().ToFloat64(),
			Low:          candle.GetLow().ToFloat64(),
		}

		// Send to all subscribers
		t.mu.RLock()
		for _, ch := range t.candleSubscribers[subscribeInfo] {
			select {
			case ch <- ourCandle:
			default:
				// Skip if channel is full
			}
		}
		t.mu.RUnlock()
	}
}

func (t *TinkoffMarketdata) convertToTinkoffCandleInterval(interval SubscriptionInterval) pb.CandleInterval {
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
