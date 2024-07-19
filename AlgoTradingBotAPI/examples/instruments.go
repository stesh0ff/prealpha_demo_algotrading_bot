package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tinkoff/invest-api-go-sdk/investgo"
	pb "github.com/tinkoff/invest-api-go-sdk/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// загружаем конфигурацию для сдк из .yaml файла
	config, err := investgo.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("config loading error %v", err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()
	// сдк использует для внутреннего логирования investgo.Logger
	// для примера передадим uber.zap
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	zapConfig.EncoderConfig.TimeKey = "time"
	l, err := zapConfig.Build()
	logger := l.Sugar()
	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Printf(err.Error())
		}
	}()
	if err != nil {
		log.Fatalf("logger creating error %v", err)
	}
	// создаем клиента для investAPI, он позволяет создавать нужные сервисы и уже
	// через них вызывать нужные методы
	client, err := investgo.NewClient(ctx, config, logger)
	if err != nil {
		logger.Fatalf("client creating error %v", err.Error())
	}
	defer func() {
		logger.Infof("closing client connection")
		err := client.Stop()
		if err != nil {
			logger.Errorf("client shutdown error %v", err.Error())
		}
	}()

	// создаем клиента для сервиса инструментов
	instrumentsService := client.NewInstrumentsServiceClient()

	instrResp, err := instrumentsService.FindInstrument("TCSG")
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		ins := instrResp.GetInstruments()
		for _, instrument := range ins {
			fmt.Printf("по запросу TCSG - %v\n", instrument.GetName())
		}
	}

	instrResp1, err := instrumentsService.FindInstrument("Тинькофф")
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		ins := instrResp1.GetInstruments()
		for _, instrument := range ins {
			fmt.Printf("по запросу Тинькофф - %v, uid - %v\n", instrument.GetName(), instrument.GetUid())
		}
	}

	scheduleResp, err := instrumentsService.TradingSchedules("MOEX", time.Now(), time.Now().Add(time.Hour*24))
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		exs := scheduleResp.GetExchanges()
		for _, ex := range exs {
			fmt.Printf("days = %v\n", ex.GetDays())
		}
	}

	// методы получения инструмента определенного типа по идентефикатору или методы получения списка
	// инструментов по торговому статусу аналогичны друг другу по входным параметрам
	sharesIds := "6e061639-6198-4448-9568-1eadb1b0e127,e2bd2eba-75de-4127-b39c-2f2dbe3866c3,ba64a3c7-dd1d-4f19-8758-94aac17d971b,3c0748ce-9b49-43e9-b788-048a6cb65174,46ae47ee-f409-4776-bf20-43a040b9e7fb,efdb54d3-2f92-44da-b7a3-8849e96039f6,7132b1c9-ee26-4464-b5b5-1046264b61d9,8e2b0325-0292-4654-8a18-4f63ed3b0e09,1c69e020-f3b1-455c-affa-45f8b8049234,77cb416f-a91e-48bd-8083-db0396c61a41,f866872b-8f68-4b6e-930f-749fe9aa79c0,cf1c6158-a303-43ac-89eb-9b1db8f96043,30817fea-20e6-4fee-ab1f-d20fc1a1bb72,21423d2d-9009-4d37-9325-883b368d13ae,88468f6c-c67a-4fb4-a006-53eed803883c,eb4ba863-e85f-4f80-8c29-f2627938ee58,02eda274-10c4-4815-8e02-a8ee7eaf485b,88e130e8-5b68-4b05-b9ae-baf32f5a3f21,9ba367af-dfbd-4d9c-8730-4b1d5a47756e,120a928b-b2d6-45d7-a445-f6e49614ae6d,2dfbc1fd-b92a-436e-b011-928c79e805f2,62560f05-3fd0-4d65-88f0-a27f249cc6de,7bedd86b-478d-4742-a28c-29d27f8dbc7d,3d8f2777-4a11-4713-af60-8038d11e66fa,2c0e65d1-d8da-477c-a438-ac41bb132e25,ebfda284-4291-4337-9dfb-f55610d0a907,03d5e771-fc10-438e-8892-85a40733612d,fd417230-19cf-4e7b-9623-f7c9ca18ec6b,fa6aae10-b8d5-48c8-bbfd-d320d925d096,"
	shareIdsMassive := strings.Split(sharesIds, ",")
	for _, shareUid := range shareIdsMassive {
		tcsResp, err := instrumentsService.ShareByUid(shareUid)
		if err != nil {
			logger.Errorf(err.Error())
		} else {
			fmt.Printf("[%v] - %v Exchange - %v, currency -  %v, ipo date - %v\n",
				tcsResp.GetInstrument().GetTicker(), tcsResp.GetInstrument().GetName(), tcsResp.GetInstrument().GetExchange(), tcsResp.GetInstrument().GetCurrency(), tcsResp.GetInstrument().GetIpoDate().AsTime().String())
		}

	}
	tcsResp, err := instrumentsService.ShareByUid("6afa6f80-03a7-4d83-9cf0-c19d7d021f76")
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		fmt.Printf("TCSG share currency -  %v, ipo date - %v\n",
			tcsResp.GetInstrument().GetCurrency(), tcsResp.GetInstrument().GetIpoDate().AsTime().String())
	}

	bondsResp, err := instrumentsService.Bonds(pb.InstrumentStatus_INSTRUMENT_STATUS_BASE)
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		bonds := bondsResp.GetInstruments()
		for i, b := range bonds {
			fmt.Printf("bond %v = %v\n", i, b.GetFigi())
			if i > 4 {
				break
			}
		}
	}

	bond, err := instrumentsService.BondByFigi("BBG00QXGFHS6")
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		fmt.Printf("bond by figi = %v\n", bond.GetInstrument().String())
	}

	interestsResp, err := instrumentsService.GetAccruedInterests("BBG00QXGFHS6", time.Now().Add(-72*time.Hour), time.Now())
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		in := interestsResp.GetAccruedInterests()
		for _, interest := range in {
			fmt.Printf("Interest = %v\n", interest.GetValue().ToFloat())
		}
	}

	bondCouponsResp, err := instrumentsService.GetBondCoupons("BBG00QXGFHS6", time.Now(), time.Now().Add(time.Hour*10000))
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		ev := bondCouponsResp.GetEvents()
		for _, coupon := range ev {
			fmt.Printf("coupon date = %v\n", coupon.GetCouponDate().AsTime().String())
		}
	}

	dividentsResp, err := instrumentsService.GetDividents("BBG004730N88", time.Now().Add(-1000*time.Hour), time.Now())
	if err != nil {
		logger.Errorf(err.Error())
		fmt.Printf("header msg = %v\n", dividentsResp.GetHeader().Get("message"))
	} else {
		divs := dividentsResp.GetDividends()
		for i, div := range divs {
			fmt.Printf("divident %v, declared date = %v\n", i, div.GetDeclaredDate().AsTime().String())
		}
	}
}
