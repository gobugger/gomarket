package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	application "github.com/gobugger/gomarket/internal/app"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/ui/templ"
	"github.com/gorilla/csrf"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func RouteMarket(app *application.Application) http.Handler {
	var port string
	parts := strings.Split(config.Addr, ":")
	if len(parts) != 2 {
		slog.Error("invalid address", "address", config.Addr)
	}
	port = parts[1]

	CSRF := csrf.Protect(
		[]byte(config.CsrfAuthKey),
		csrf.Secure(false),
		csrf.TrustedOrigins([]string{"localhost:" + port, config.OnionAddr}),
		csrf.FieldName("csrf.Token"),
		csrf.CookieName("_csrf"),
		csrf.Path("/"),
	)

	r := chi.NewRouter()
	r.Use(application.LimitBodySize(32<<20),
		middleware.ThrottleBacklog(400, 800, time.Second),
		middleware.Recoverer,
		app.SessionManager.LoadAndSave,
		CSRF,
		application.SetSecureHeaders,
		app.InjectLogger,
		app.LogRequest,
		middleware.Compress(5))

	r.Group(func(r chi.Router) {
		r.Get("/static/{filepath:.*}", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))).ServeHTTP)
		r.Get("/ui/css/{filepath:.*}", http.StripPrefix("/ui/css/", http.FileServer(http.Dir("./ui/css"))).ServeHTTP)
		r.Get("/", app.Index)
		if config.EntryGuardEnabled() {
			r.Get("/jail", app.Jail)
			r.Post("/jail", app.HandleJail)
		}
	})

	r.Group(func(r chi.Router) {
		r.Use(httprate.Limit(
			30,
			time.Second,
			httprate.WithKeyFuncs(app.GetRateLimitKey),
			httprate.WithResponseHeaders(httprate.ResponseHeaders{}),
			httprate.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, _ error) {
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}),
		))

		r.Get("/uploads/*", http.StripPrefix("/uploads/", http.HandlerFunc(app.ServeUpload)).ServeHTTP)
		r.Get("/pgp.txt", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(config.PgpKey)) })    // nolint
		r.Get("/canary.txt", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(config.Canary)) }) // nolint
		r.Get("/support", app.ServeTemplate(ui.Support))
		r.Get("/support/faq", app.Faq)
		r.Get("/products", app.Products)

		r.Group(func(r chi.Router) {
			r.Use(app.RequireNoAuth)

			r.Get("/register", app.Register)
			r.Get("/login", app.Login)
			r.Post("/register", app.HandleRegister)
			r.Post("/login", app.HandleLogin)
			r.Post("/login/2fa", app.Handle2FA)
		})

		r.Group(func(r chi.Router) {
			r.Use(app.RequireAuth)

			r.Get("/product", app.Product)
			r.Get("/order", app.ViewOrder)
			r.Get("/orders/all", app.Orders)
			r.Get("/orders/review", app.Review)
			r.Get("/license/application", app.VendorApplication)
			r.Get("/settings", app.Settings)
			r.Get("/wallet", app.Wallet)
			r.Get("/cart", app.Cart)
			r.Get("/checkout", app.Checkout)
			r.Get("/support/ticket/create", app.CreateTicket)
			r.Get("/support/tickets", app.Tickets)
			r.Get("/support/ticket", app.Ticket)
			r.Get("/vendor", app.Vendor)
			r.Get("/notifications", app.Notifications)
			r.Post("/logout", app.HandleLogout)
			r.Post("/product/action", app.HandleProductAction)
			r.Post("/cart/update", app.HandleUpdateCart)
			r.Post("/cart/delete", app.HandleDeleteCart)
			r.Post("/checkout/order", app.HandleOrder)
			r.Post("/orders/cancel", app.HandleCancelOrder)
			r.Post("/orders/review", app.HandleReview)
			r.Post("/orders/dispute", app.HandleDispute)
			r.Post("/orders/extend", app.HandleExtend)
			r.Post("/order/chat", app.HandleOrderChat)
			r.Post("/order/dispute/offer/response", app.HandleDisputeOfferResponse)
			r.Post("/wallet/withdraw", app.HandleWithdrawal)
			r.Post("/license/application", app.HandleVendorApplication)
			r.Post("/settings/update", app.HandleUpdateSettings)
			r.Post("/settings/change-password", app.HandleChangePassword)
			r.Post("/settings/updatepgp", app.HandleUpdatePGP)
			r.Post("/settings/checkpgp", app.HandleCheckPGP)
			r.Post("/support/ticket/create", app.HandleCreateTicket)
			r.Post("/support/ticket/response", app.HandleTicketResponse)
			r.Post("/notification/delete", app.HandleDeleteNotification)
			r.Post("/notification/delete/all", app.HandleDeleteAllNotifications)
		})

		r.Group(func(r chi.Router) {
			r.Use(app.RequireVendor)

			r.Get("/create-listing", app.ViewCreateListing)
			r.Post("/create-listing", app.HandleCreateListing)
			r.Post("/orders/dispute/offer", app.HandleDisputeOffer)
			r.Post("/orders/process", app.HandleProcessOrder)
			r.Post("/orders/deliver", app.HandleDeliver)
			r.Post("/product/update", app.HandleUpdateProduct)
			r.Post("/settings/profile", app.HandleUpdateVendorProfile)
			r.Post("/settings/delivery_methods", app.HandleUpdateDeliveryMethods)
		})
	})

	return r
}
