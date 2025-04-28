package main

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

func main() {
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

	instrumentsClient := client.NewInstrumentsServiceClient()

	instrument, err := instrumentsClient.ShareByTicker("VKCO", "TQBR")
	if err != nil {
		panic(err)
	}

	fmt.Println(instrument)

	mdStreamClient := client.NewMarketDataStreamClient()

	// для синхронизации всех горутин
	wg := &sync.WaitGroup{}

	// создаем стримов сколько нужно, например 2
	firstMDStream, err := mdStreamClient.MarketDataStream()
	if err != nil {
		logger.Errorf(err.Error())
	}
	// результат подписки на инструменты это канал с определенным типом информации, при повторном вызове функции
	// подписки(например на свечи), возвращаемый канал можно игнорировать, так как при первом вызове он уже был получен
	firstInstrumetsGroup := []string{"BBG004730N88"}
	candleChan, err := firstMDStream.SubscribeCandle(
		firstInstrumetsGroup,
		pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE,
		true,
		nil,
	)
	if err != nil {
		logger.Errorf(err.Error())
	}

	// secondInstrumetsGroup := []string{"BBG00475KKY8", "BBG004RVFCY3"}

	// lastPriceChan, err := firstMDStream.SubscribeLastPrice(
	// 	secondInstrumetsGroup,
	// )
	// if err != nil {
	// 	logger.Errorf(err.Error())
	// }

	// firstMDStream.UnSubscribeCandle(firstInstrumetsGroup, pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE, true, nil)

	// функцию Listen нужно вызвать один раз для каждого стрима и в отдельной горутине
	// для остановки стрима можно использовать метод Stop, он отменяет контекст внутри стрима
	// после вызова Stop закрываются каналы и завершается функция Listen
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := firstMDStream.Listen()
		if err != nil {
			logger.Errorf(err.Error())
		}
	}()

	// для дальнейшей обработки, поступившей из канала, информации хорошо подойдет механизм,
	// основанный на паттерне pipeline https://go.dev/blog/pipelines

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				logger.Infof("stop listening first channels")
				return
			case candle, ok := <-candleChan:
				if !ok {
					return
				}
				// клиентская логика обработки...
				fmt.Println("Candle high price = ", candle.GetHigh().ToFloat())
				// case price, ok := <-lastPriceChan:
				// 	if !ok {
				// 		return
				// 	}
				// 	// клиентская логика обработки...
				// 	fmt.Println("Price high price = ", price.GetLastPriceType())
			}
		}
	}(ctx)

	wg.Wait()
}
