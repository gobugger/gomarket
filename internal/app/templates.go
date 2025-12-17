package app

import (
	"github.com/gobugger/globalize"
	"github.com/gobugger/gomarket/internal/captcha"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/vendor"
	ui "github.com/gobugger/gomarket/ui/templ"
	"github.com/gorilla/csrf"
	"log/slog"
	"net/http"
)

func (app *Application) newTemplateContext(req *http.Request) (*ui.TemplateContext, error) {
	ctx := req.Context()
	q := repo.New(app.Db)

	tc := &ui.TemplateContext{}

	user := app.loggedInUser(ctx)
	if user != nil {
		tc.UID = user.ID
		tc.AuthLevel = ui.AuthLevelCustomer
		tc.Username = user.Username
		tc.PgpKey = user.PgpKey
		tc.Settings.TwofaEnabled = user.TwofaEnabled

		wallet, err := q.GetWalletForUser(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		tc.BalancePico = wallet.BalancePico

		if vendor.HasLicense(ctx, q, user.ID) {
			tc.AuthLevel = ui.AuthLevelVendor
		}

		numUnseenNotifications, err := q.GetCountOfUnseenNotificationsForUser(ctx, user.ID)
		if err != nil {
			getLoggerFromRequest(req).Error("number of unseen notifications", slog.Any("error", err))
		}
		tc.NumUnseenNotifications = int(numUnseenNotifications)

		n, err := q.GetNumCartItemsForCustomer(ctx, user.ID)
		if err != nil {
			getLoggerFromRequest(req).Error("number of cart items", slog.Any("error", err))
		}
		tc.NumCartItems = int(n)
	}

	tc.Settings.Lang = func() string {
		locale := app.SessionManager.GetString(ctx, localeKey)
		if locale == "" {
			if user != nil {
				locale = user.Locale
			} else {
				locale = globalize.DefaultLocale
			}
			app.SessionManager.Put(ctx, localeKey, locale)
			return locale
		}
		return locale
	}()

	tc.Settings.Currency = func() string {
		if user != nil {
			return user.Currency
		} else {
			return string(currency.DefaultCurrency)
		}
	}()

	numUsers, _ := q.GetNumberOfUsers(ctx)
	tc.Stats.NumUsers = numUsers

	numVendors, _ := q.GetNumberOfVendorLicenses(ctx)
	tc.Stats.NumVendors = numVendors

	notes, _ := app.SessionManager.Pop(ctx, uiNotesKey).([]ui.Note)
	app.SessionManager.Pop(ctx, notesKey)
	tc.Notes = notes

	tc.Config.Name = config.SiteName
	tc.Config.Onion = config.OnionAddr
	tc.CsrfField = string(csrf.TemplateField(req))
	tc.CaptchaSrc = captcha.TemplFieldSrc(ctx, app.SessionManager)
	tc.Form = popForm(ctx, app.SessionManager)

	return tc, nil
}
