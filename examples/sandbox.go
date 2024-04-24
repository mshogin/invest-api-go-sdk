package main

import (
	"context"
	"fmt"
	"github.com/tinkoff/invest-api-go-sdk/investgo"
	pb "github.com/tinkoff/invest-api-go-sdk/proto"
	"go.uber.org/zap"
	"log"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	config, err := investgo.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("config loading error %v", err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

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

	sandboxService := client.NewSandboxServiceClient()

	var newAccId string

	accountsResp, err := sandboxService.GetSandboxAccounts()
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		accs := accountsResp.GetAccounts()
		if len(accs) > 0 {
			newAccId = accs[0].GetId()
		} else {
			openAccount, err := sandboxService.OpenSandboxAccount()
			if err != nil {
				logger.Errorf(err.Error())
			} else {
				newAccId = openAccount.GetAccountId()
			}
			client.Config.AccountId = newAccId
		}
	}
	payInResp, err := sandboxService.SandboxPayIn(&investgo.SandboxPayInRequest{
		AccountId: newAccId,
		Currency:  "RUB",
		Unit:      100000,
		Nano:      0,
	})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		fmt.Printf("sandbox accouunt %v balance = %v\n", newAccId, payInResp.GetBalance().ToFloat())
	}

	instrumentsService := client.NewInstrumentsServiceClient()

	var id string
	instrumentResp, err := instrumentsService.FindInstrument("TCSG")
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		instruments := instrumentResp.GetInstruments()
		for _, instrument := range instruments {
			if strings.Compare(instrument.GetTicker(), "TCSG") == 0 {
				id = instrument.GetUid()
			}
		}
	}
	ordersService := client.NewOrdersServiceClient()

	buyResp, err := ordersService.Buy(&investgo.PostOrderRequestShort{
		InstrumentId: id,
		Quantity:     1,
		Price:        nil,
		AccountId:    newAccId,
		OrderType:    pb.OrderType_ORDER_TYPE_MARKET,
		OrderId:      investgo.CreateUid(),
	})
	if err != nil {
		logger.Errorf(err.Error())
		fmt.Printf("msg = %v\n", investgo.MessageFromHeader(buyResp.GetHeader()))
	} else {
		fmt.Printf("order status = %v\n", buyResp.GetExecutionReportStatus().String())
	}

	operationsService := client.NewOperationsServiceClient()

	positionsResp, err := operationsService.GetPositions(newAccId)
	if err != nil {
		logger.Errorf(err.Error())
		fmt.Printf("msg = %v\n", investgo.MessageFromHeader(buyResp.GetHeader()))
	} else {
		positions := positionsResp.GetSecurities()
		for i, position := range positions {
			fmt.Printf("position number %v, uid = %v\n", i, position.GetInstrumentUid())
		}
	}

	sellResp, err := ordersService.Sell(&investgo.PostOrderRequestShort{
		InstrumentId: id,
		Quantity:     1,
		Price:        nil,
		AccountId:    newAccId,
		OrderType:    pb.OrderType_ORDER_TYPE_MARKET,
		OrderId:      investgo.CreateUid(),
	})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		fmt.Printf("order status = %v\n", sellResp.GetExecutionReportStatus().String())
	}

}
