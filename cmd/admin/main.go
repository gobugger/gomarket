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
	"github.com/gobugger/gomarket/internal/util/db"
	"github.com/gobugger/gomarket/internal/util/uow"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log/slog"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	config.ParseAndLoad()

	db, err := db.Open(context.Background(), config.DSN)
	if err != nil {
		slog.Error("open db", slog.Any("error", err))
	}
	defer db.Close()

	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(db)

	minioClient, err := minio.New(config.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Minio.AccessKeyID, config.Minio.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		slog.Error("new minio client", slog.Any("error", err))
	}

	app := app.Application{
		Db:             db,
		SessionManager: sessionManager,
		MinioClient:    minioClient,
		UoW:            uow.New(db),
	}

	srv := http.Server{
		Addr:              config.Addr,
		Handler:           route.RouteAdmin(&app),
		ReadHeaderTimeout: time.Second * 10,
	}

	var wg sync.WaitGroup
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := currency.Start(ctx, repo.New(db), 15*time.Minute); err != nil {
		slog.Error("Failed to start currency service", slog.Any("error", err))
	}

	wg.Add(1)
	go func() {
		slog.Info("server started", "addr", config.Addr)
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
