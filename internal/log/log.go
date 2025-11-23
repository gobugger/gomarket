package log

import (
	"context"
	"log/slog"
	"os"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func Set(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, "logger", l)
}

func Get(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value("logger").(*slog.Logger)
	if ok {
		return l
	} else {
		return slog.Default()
	}
}
