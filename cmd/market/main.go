package main

import (
	"context"
	"errors"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gobugger/gomarket/internal/app"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/route"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/payment"
	"github.com/gobugger/gomarket/internal/util"
	"github.com/gobugger/gomarket/internal/util/db"
	"github.com/gobugger/gomarket/internal/util/uow"
	"github.com/gobugger/gomarket/internal/worker"
	"github.com/gobugger/gomarket/pkg/jail"
	"github.com/gobugger/gomarket/pkg/payment/processor"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func run(ctx context.Context, app *app.Application) {
	srv := http.Server{
		Addr:              config.Addr,
		Handler:           route.RouteMarket(app),
		ReadHeaderTimeout: 30 * time.Second,
	}

	var wg sync.WaitGroup
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := currency.Start(ctx, repo.New(app.Db), 15*time.Minute); err != nil {
		slog.Error("start currency package", slog.Any("error", err))
		return
	}

	if err := app.RiverClient.Start(ctx); err != nil {
		slog.Error("failed to start river client", slog.Any("error", err))
		return
	}

	// Run server
	wg.Add(1)
	go func() {
		slog.Info("starting server", "onionAddr", config.OnionAddr, "addr", config.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed", slog.Any("error", err))
			cancel()
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		<-ctx.Done()

		slog.Debug("Shutting down servers")
		if err := srv.Shutdown(context.TODO()); err != nil {
			slog.Error("server failed to shutdown", slog.Any("error", err))
		}

		wg.Done()
	}()

	wg.Wait()
}

func main() {
	config.ParseAndLoad()

	ctx := context.Background()

	db, err := db.Open(ctx, config.DSN)
	if err != nil {
		slog.Error("connect to db", "dsn", config.DSN, slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	sessionManager := scs.New()
	sessionManager.Lifetime = 6 * time.Hour
	sessionManager.Store = pgxstore.New(db)

	minioClient, err := minio.New(config.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Minio.AccessKeyID, config.Minio.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		slog.Error("new minio client", slog.Any("error", err))
		os.Exit(1)
	}

	buckets := []string{
		"product",
		"thumbnail",
		"logo",
		"inventory",
	}

	for _, bucket := range buckets {
		if ok, _ := minioClient.BucketExists(ctx, bucket); !ok {
			if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				slog.Error("failed to make bucket", slog.String("bucket", bucket), slog.Any("error", err))
				os.Exit(1)
			}
		}
	}

	var paymentProcessor processor.Processor
	if config.DevMode() {
		paymentProcessor = payment.NewFakeClient()
	} else {
		paymentProcessor = processor.NewMoneropayClient(config.MoneropayURL)
	}

	if err := util.RiverMigrate(ctx, db); err != nil {
		slog.Error("failed to migrate river", slog.Any("error", err))
		os.Exit(1)
	}

	workers := river.NewWorkers()
	river.AddWorker(workers, &worker.PrepareInvoiceWorker{
		Db: db,
		Pp: paymentProcessor,
	})

	rc, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		slog.Error("failed to create river client", slog.Any("error", err))
		os.Exit(1)
	}

	app := app.Application{
		Db:               db,
		SessionManager:   sessionManager,
		MinioClient:      minioClient,
		PaymentProcessor: paymentProcessor,
		RiverClient:      rc,
		UoW:              uow.New(db),
	}

	if config.EntryGuardEnabled() {
		go jail.SetupAndStart()
	}

	run(ctx, &app)
}
