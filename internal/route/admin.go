package route

import (
	application "github.com/gobugger/gomarket/internal/app"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gorilla/csrf"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"log/slog"
	"net/http"
	"strings"
)

func RouteAdmin(app *application.Application) http.Handler {
	r := httprouter.New()

	staticServer := http.FileServer(http.Dir(config.StaticDir))
	r.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static/", staticServer))

	cssServer := http.FileServer(http.Dir(config.CssDir))
	r.Handler(http.MethodGet, "/ui/css/*filepath", http.StripPrefix("/ui/css/", cssServer))

	r.Handler(http.MethodGet, "/uploads/*filepath", http.StripPrefix("/uploads/", http.HandlerFunc(app.ServeUpload)))

	r.HandlerFunc(http.MethodGet, "/", app.Admin)
	r.HandlerFunc(http.MethodGet, "/health", app.Health)
	r.HandlerFunc(http.MethodGet, "/application", app.AdminVendorApplication)
	r.HandlerFunc(http.MethodGet, "/ticket", app.AdminTicket)

	r.HandlerFunc(http.MethodPost, "/delete", app.AdminHandleOperation)
	r.HandlerFunc(http.MethodPost, "/application", app.AdminHandleVendorApplication)
	r.HandlerFunc(http.MethodPost, "/ticket", app.AdminHandleTicket)
	r.HandlerFunc(http.MethodPost, "/settings", app.AdminHandleSettings)
	r.HandlerFunc(http.MethodPost, "/category/add", app.AdminHandleAddCategory)

	var port string
	parts := strings.Split(config.Addr, ":")
	if len(parts) != 2 {
		slog.Error("invalid address", "address", config.Addr)
	}
	port = parts[1]

	CSRF := csrf.Protect([]byte(config.CsrfAuthKey),
		csrf.Path("/"),
		csrf.Secure(false),
		csrf.TrustedOrigins([]string{"localhost:" + port, config.OnionAddr}),
		csrf.FieldName("csrf.Token"),
		csrf.CookieName("_csrf"),
	)

	secure := alice.New(
		app.SessionManager.LoadAndSave,
		CSRF,
		application.SetSecureHeaders,
		app.LogRequest)

	return secure.Then(r)
}
