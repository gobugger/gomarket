package main

import (
	"context"
	"github.com/go-co-op/gocron/v2"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/order"
	"github.com/gobugger/gomarket/internal/service/payment"
	"github.com/gobugger/gomarket/internal/service/payment/wallet"
	"github.com/gobugger/gomarket/internal/util/db"
	"github.com/gobugger/gomarket/internal/util/uow"
	"github.com/gobugger/gomarket/pkg/payment/processor"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/net/proxy"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var ready = atomic.Bool{}

func makeClient() (*http.Client, error) {
	if config.Socks5Hostname == "" {
		slog.Warn("Worker is not using proxy")
		return &http.Client{}, nil
	}

	dialer, err := proxy.SOCKS5("tcp", config.Socks5Hostname, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{DialContext: dialer.(proxy.ContextDialer).DialContext},
		Timeout:   20 * time.Second,
	}, nil
}

func main() {
	config.ParseAndLoad()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := db.Open(ctx, config.DSN)
	if err != nil {
		slog.Error("Failed to connect to db", "dsn", config.DSN, slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	var pp processor.Processor
	if config.DevMode() {
		pp = payment.NewFakeClient()
	} else {
		pp = processor.NewMoneropayClient(config.MoneropayURL)
	}

	client, err := makeClient()
	if err != nil {
		slog.Error("Failed to initialize client", slog.Any("error", err))
		os.Exit(1)
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("Failed to create scheduler", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() { s.Shutdown() }()

	if err := registerJobs(ctx, s, db, pp, client); err != nil {
		slog.Error("Failed to register jobs", slog.Any("error", err))
		os.Exit(1)
	}

	s.Start()

	go func() {
		http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
			if ready.Load() {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		})

		if err := http.ListenAndServe(config.Addr, http.DefaultServeMux); err != nil {
			slog.Error("Server failed", slog.Any("error", err))
		}
	}()

	<-ctx.Done()

}

var orderJobs = []struct {
	job    func(ctx context.Context, qtx *repo.Queries) error
	errMsg string
}{
	{
		job: func(ctx context.Context, qtx *repo.Queries) error {
			return order.ProcessPaid(ctx, qtx)
		},
		errMsg: "Failed to process paid orders",
	},
	{
		job: func(ctx context.Context, qtx *repo.Queries) error {
			return order.AutoFinalize(ctx, qtx, config.OrderDeliveryWindow)
		},
		errMsg: "Failed to auto finalize orders",
	},
	{
		job: func(ctx context.Context, qtx *repo.Queries) error {
			return order.DeclineUnhandled(ctx, qtx, config.OrderProcessingWindow, config.OrderDispatchWindow)
		},
		errMsg: "Failed to decline unhandled orders",
	},
	{
		job: func(ctx context.Context, qtx *repo.Queries) error {
			return order.CancelExpired(ctx, qtx)
		},
		errMsg: "Failed to cancel expired orders",
	},
}

func registerJobs(ctx context.Context, s gocron.Scheduler, db *pgxpool.Pool, pp processor.Processor, client *http.Client) error {
	uow := uow.New(db)

	_, err := s.NewJob(gocron.DurationJob(15*time.Minute),
		gocron.NewTask(func(ctx context.Context, c *http.Client, url string) {
			err := updateXMRPrice(ctx, db, c, url)
			if err != nil {
				slog.Error("Failed to update XMR price", slog.Any("error", err))
			} else {
				ready.Store(true)
			}
		}, client, cryptocompareURL()),
		gocron.WithContext(ctx),
		gocron.WithStartAt(gocron.WithStartImmediately()),
	)
	if err != nil {
		return err
	}

	_, err = s.NewJob(gocron.DurationJob(time.Minute),
		gocron.NewTask(func(ctx context.Context) {
			err := uow.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
				return payment.ProcessInvoices(ctx, qtx, pp, config.InvoicePaymentWindow)
			})
			if err != nil {
				slog.Error("Failed to process invoices", slog.Any("error", err))
			}
		}),
		gocron.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	_, err = s.NewJob(gocron.DurationJob(time.Minute),
		gocron.NewTask(func(ctx context.Context) {
			err := uow.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
				return wallet.HandleDeposits(ctx, qtx)
			})
			if err != nil {
				slog.Error("Failed to handle deposits", slog.Any("error", err))
			}
		}),
		gocron.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	_, err = s.NewJob(gocron.DurationJob(time.Minute),
		gocron.NewTask(func(ctx context.Context) {
			err := uow.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
				return payment.HandleWithdrawals(ctx, qtx, pp)
			})
			if err != nil {
				slog.Error("Failed to handle withdrawals", slog.Any("error", err))
			}
		}),
		gocron.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	_, err = s.NewJob(gocron.DurationJob(time.Minute),
		gocron.NewTask(func(ctx context.Context) {
			for _, oj := range orderJobs {
				err := uow.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
					return oj.job(ctx, qtx)
				})
				if err != nil {
					slog.Error(oj.errMsg, slog.Any("error", err))
				}
			}
		}),
		gocron.WithContext(ctx),
	)

	return err
}
