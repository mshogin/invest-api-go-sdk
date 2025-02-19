package investgo

import (
	"context"
	"crypto/tls"
	"fmt"
	pb "github.com/tinkoff/invest-api-go-sdk/proto"
	"github.com/tinkoff/invest-api-go-sdk/retry"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
	"strings"
	"time"
)

type ctxKey string

type Client struct {
	conn   *grpc.ClientConn
	Config Config
	Logger Logger
	ctx    context.Context
}

// NewClient - создание клиента для API Тинькофф инвестиций
func NewClient(ctx context.Context, conf Config, l Logger) (*Client, error) {
	setDefaultConfig(&conf)

	var authKey ctxKey = "authorization"
	ctx = context.WithValue(ctx, authKey, fmt.Sprintf("Bearer %s", conf.Token))
	ctx = metadata.AppendToOutgoingContext(ctx, "x-app-name", conf.AppName)

	Opts := []retry.CallOption{
		retry.WithCodes(codes.Unavailable, codes.Internal),
		retry.WithBackoff(retry.BackoffLinear(500 * time.Millisecond)),
		retry.WithMax(conf.MaxRetries),
	}

	// при исчерпывании лимита запросов в минуту, нужно ждать дольше
	exhaustedOpts := []retry.CallOption{
		retry.WithCodes(codes.ResourceExhausted),
		retry.WithMax(conf.MaxRetries),
	}

	streamInterceptors := []grpc.StreamClientInterceptor{
		retry.StreamClientInterceptor(Opts...),
	}

	var unaryInterceptors []grpc.UnaryClientInterceptor
	if conf.DisableResourceExhaustedRetry {
		unaryInterceptors = []grpc.UnaryClientInterceptor{
			retry.UnaryClientInterceptor(Opts...),
		}
	} else {
		unaryInterceptors = []grpc.UnaryClientInterceptor{
			retry.UnaryClientInterceptor(Opts...),
			retry.UnaryClientInterceptorRE(exhaustedOpts...),
		}
	}

	conn, err := grpc.Dial(conf.EndPoint,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(oauth.TokenSource{
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Token}),
		}),
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
		grpc.WithChainStreamInterceptor(streamInterceptors...))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		Config: conf,
		Logger: l,
		ctx:    ctx,
	}, nil
}

func setDefaultConfig(conf *Config) {
	if strings.Compare(conf.AppName, "") == 0 {
		conf.AppName = "invest-api-go-sdk"
	}
	if strings.Compare(conf.EndPoint, "") == 0 {
		conf.EndPoint = "sandbox-invest-public-api.tinkoff.ru:443"
	}
	if conf.DisableAllRetry {
		conf.MaxRetries = 0
	} else if conf.MaxRetries == 0 {
		conf.MaxRetries = 3
	}
}

type Logger interface {
	Infof(template string, args ...any)
	Errorf(template string, args ...any)
	Fatalf(template string, args ...any)
}

// NewMarketDataStreamClient - создание клиента для сервиса стримов маркетадаты
func (c *Client) NewMarketDataStreamClient() *MarketDataStreamClient {
	pbClient := pb.NewMarketDataStreamServiceClient(c.conn)
	return &MarketDataStreamClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// NewOrdersServiceClient - создание клиента сервиса ордеров
func (c *Client) NewOrdersServiceClient() *OrdersServiceClient {
	pbClient := pb.NewOrdersServiceClient(c.conn)
	return &OrdersServiceClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// NewMarketDataServiceClient - создание клиента сервиса маркетдаты
func (c *Client) NewMarketDataServiceClient() *MarketDataServiceClient {
	pbClient := pb.NewMarketDataServiceClient(c.conn)
	return &MarketDataServiceClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// NewInstrumentsServiceClient - создание клиента сервиса инструментов
func (c *Client) NewInstrumentsServiceClient() *InstrumentsServiceClient {
	pbClient := pb.NewInstrumentsServiceClient(c.conn)
	return &InstrumentsServiceClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// NewUsersServiceClient - создание клиента сервиса счетов
func (c *Client) NewUsersServiceClient() *UsersServiceClient {
	pbClient := pb.NewUsersServiceClient(c.conn)
	return &UsersServiceClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// NewOperationsServiceClient - создание клиента сервиса операций
func (c *Client) NewOperationsServiceClient() *OperationsServiceClient {
	pbClient := pb.NewOperationsServiceClient(c.conn)
	return &OperationsServiceClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// NewStopOrdersServiceClient - создание клиента сервиса стоп-ордеров
func (c *Client) NewStopOrdersServiceClient() *StopOrdersServiceClient {
	pbClient := pb.NewStopOrdersServiceClient(c.conn)
	return &StopOrdersServiceClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// NewSandboxServiceClient - создание клиента для работы с песочницей
func (c *Client) NewSandboxServiceClient() *SandboxServiceClient {
	pbClient := pb.NewSandboxServiceClient(c.conn)
	return &SandboxServiceClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// NewOrdersStreamClient - создание клиента стримов сделок
func (c *Client) NewOrdersStreamClient() *OrdersStreamClient {
	pbClient := pb.NewOrdersStreamServiceClient(c.conn)
	return &OrdersStreamClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// NewOperationsStreamClient - создание клиента стримов обновлений портфеля
func (c *Client) NewOperationsStreamClient() *OperationsStreamClient {
	pbClient := pb.NewOperationsStreamServiceClient(c.conn)
	return &OperationsStreamClient{
		conn:     c.conn,
		config:   c.Config,
		logger:   c.Logger,
		ctx:      c.ctx,
		pbClient: pbClient,
	}
}

// Stop - корректное завершение работы клиента
func (c *Client) Stop() error {
	// TODO add some stop options
	c.Logger.Infof("stop client")
	return c.conn.Close()
}
