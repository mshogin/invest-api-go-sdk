package main

import (
	"context"
	"fmt"
	"github.com/tinkoff/invest-api-go-sdk/investgo"
	pb "github.com/tinkoff/invest-api-go-sdk/proto"
	"go.uber.org/zap"
	"log"
	"os/signal"
	"syscall"
	"time"
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
	prod := zap.NewExample()
	defer func() {
		err := prod.Sync()
		if err != nil {
			log.Printf("Prod.Sync %v", err.Error())
		}
	}()
	if err != nil {
		log.Fatalf("logger creating error %v", err)
	}
	logger := prod.Sugar()
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
