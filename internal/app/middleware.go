package app

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/vendor"
	"log/slog"
	"net/http"
)

func SetSecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:;")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header()["Date"] = nil

		next.ServeHTTP(w, r)
	})
}

func (app *Application) InjectLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		requestID := uuid.New()
		userID, ok := app.SessionManager.Get(ctx, userIDKey).(uuid.UUID)

		logger := slog.Default()
		if ok {
			logger = logger.With("requestID", requestID, "userID", userID)
		} else {
			logger = logger.With("requestID", requestID)
		}

		ctx = context.WithValue(ctx, "logger", logger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("reguest", "addr", r.RemoteAddr, "protocol", r.Proto, "method", r.Method, "url", r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *Application) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.SessionManager.Exists(r.Context(), userIDKey) {
			next.ServeHTTP(w, r)
		} else {
			l := app.getLocalizer(r.Context())
			app.notifyUser(r.Context(), true, l.Translate("You need to be logged in to access that page"))
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})
}

func (app *Application) RequireNoAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.SessionManager.Exists(r.Context(), userIDKey) {
			next.ServeHTTP(w, r)
		} else {
			app.clientError(w, http.StatusBadRequest)
		}
	})
}

func (app *Application) CustomerOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if uid, ok := app.SessionManager.Get(ctx, userIDKey).(uuid.UUID); ok && !vendor.HasLicense(ctx, repo.New(app.Db), uid) {
			next.ServeHTTP(w, r)
		} else {
			app.clientError(w, http.StatusUnauthorized)
		}
	})
}

func (app *Application) RequireVendor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if uid, ok := app.SessionManager.Get(ctx, userIDKey).(uuid.UUID); ok && vendor.HasLicense(ctx, repo.New(app.Db), uid) {
			next.ServeHTTP(w, r)
		} else {
			app.clientError(w, http.StatusUnauthorized)
		}
	})
}

func (app *Application) GetRateLimitKey(r *http.Request) (string, error) {
	uid, ok := app.SessionManager.Get(r.Context(), rateLimitKey).(uuid.UUID)
	if !ok {
		return "", fmt.Errorf("you need to go through the gate in order to access this resource")
	}
	return uid.String(), nil
}

func LimitBodySize(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}
}
