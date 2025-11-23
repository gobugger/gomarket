package app

import (
	"context"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/captcha"
	"github.com/gobugger/gomarket/internal/form"
	"github.com/gobugger/gomarket/internal/localizer"
	"github.com/gobugger/gomarket/internal/log"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/ui/templ"
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
)

const (
	rateLimitKey           = "rate_limit"
	userIDKey              = "userID"
	notesKey               = "notes"
	uiNotesKey             = "ui_notes"
	formKey                = "form"
	localeKey              = "locale"
	captchaSolutionKey     = "captcha_solution"
	twoFactorTokenKey      = "2fa_token"
	twoFactorUIDKey        = "2fa_uid"
	jailSolutionIndexesKey = "jail_solution"
	registerFormKey        = "register_form"
	pgpKeyKey              = "pgp_key"
)

func (app *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	logger := getLoggerFromRequest(r)
	logger.Error("server error", "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *Application) redirectBack(w http.ResponseWriter, r *http.Request) {
	if res, err := url.Parse(r.Referer()); err == nil {
		http.Redirect(w, r, res.RequestURI(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func parseID(r *http.Request) (uuid.UUID, error) {
	uuid, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		getLoggerFromRequest(r).Error("failed to parse id from query", slog.Any("error", err))
	}
	return uuid, err
}

func (app *Application) getLocalizer(ctx context.Context) localizer.Localizer {
	l, _ := localizer.Get(app.SessionManager.GetString(ctx, localeKey))
	return l
}

func (app *Application) loggedInUser(ctx context.Context) *repo.User {
	if uid, ok := app.SessionManager.Get(ctx, userIDKey).(uuid.UUID); ok {
		queries := repo.New(app.Db)
		user, err := queries.GetUser(ctx, uid)
		if err != nil {
			return nil
		}
		return &user
	}

	return nil
}

func (app *Application) notifyUser(ctx context.Context, isError bool, messages ...string) {
	tmp, ok := app.SessionManager.Get(ctx, uiNotesKey).([]ui.Note)
	if !ok {
		tmp = []ui.Note{}
	}

	i := len(tmp)
	for _, msg := range messages {
		tmp = append(tmp, ui.Note{Index: i, Message: msg, IsError: isError})
		i++
	}

	app.SessionManager.Put(ctx, uiNotesKey, tmp)
}

func (app *Application) login(ctx context.Context, uid uuid.UUID) {
	logger := log.Get(ctx)
	if err := app.SessionManager.RenewToken(ctx); err != nil {
		logger.Error("failed to renew session token", slog.Any("error", err))
	}
	app.SessionManager.Put(ctx, userIDKey, uid)
	queries := repo.New(app.Db)
	u, err := queries.UpdateUserPrevLogin(ctx, uid)
	if err != nil {
		logger.Error("failed to update previous login time", slog.Any("error", err))
		return
	}
	app.SessionManager.Put(ctx, localeKey, u.Locale)
}

func (app *Application) logout(ctx context.Context) error {
	if err := app.SessionManager.RenewToken(ctx); err != nil {
		return err
	}
	app.SessionManager.Pop(ctx, userIDKey)
	app.SessionManager.Pop(ctx, rateLimitKey)
	return nil
}

func (app *Application) Repo() *repo.Queries {
	return repo.New(app.Db)
}

func processForm[T any](r *http.Request, sessionManager *scs.SessionManager, fileRules map[string]form.FileRules) (*form.FormData[T], error) {
	fh := form.NewFormHandler[T]()
	if fileRules != nil {
		fh = fh.WithFileRules(fileRules)
	}

	var cs *captcha.Solution
	if solution, ok := sessionManager.Get(r.Context(), captcha.SolutionKey).(captcha.Solution); ok {
		cs = &solution
	}

	return fh.Parse(r, cs)
}

func putForm[T any](ctx context.Context, sessionManager *scs.SessionManager, fd *form.FormData[T]) {
	sessionManager.Put(ctx, "uiForm", ui.Form{
		Values:      fd.Values,
		FieldErrors: fd.Errors,
	})
}

func popForm(ctx context.Context, sessionManager *scs.SessionManager) ui.Form {
	form, _ := sessionManager.Pop(ctx, "uiForm").(ui.Form)
	return form
}

func getLoggerFromRequest(req *http.Request) *slog.Logger {
	return log.Get(req.Context())
}
