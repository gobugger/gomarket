package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/form"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/auth"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/market"
	"github.com/gobugger/gomarket/internal/service/notification"
	"github.com/gobugger/gomarket/internal/service/order"
	"github.com/gobugger/gomarket/internal/service/payment"
	"github.com/gobugger/gomarket/internal/service/product"
	"github.com/gobugger/gomarket/internal/service/settings"
	user_settings "github.com/gobugger/gomarket/internal/service/user"
	"github.com/gobugger/gomarket/internal/service/vendor"
	"github.com/gobugger/gomarket/internal/support"
	"github.com/gobugger/gomarket/internal/util"
	"github.com/gobugger/gomarket/internal/view"
	"github.com/gobugger/gomarket/pkg/jail"
	"github.com/gobugger/gomarket/pkg/model"
	"github.com/gobugger/gomarket/ui/templ"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"io"
	"net/http"
	"path"
	"slices"
	"strconv"
	"time"
)

func (app *Application) ServeTemplate(t func(tc *ui.TemplateContext) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tc, err := app.newTemplateContext(r)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		t(tc).Render(r.Context(), w)
	}
}

func (app *Application) Index(w http.ResponseWriter, r *http.Request) {
	if _, ok := app.SessionManager.Get(r.Context(), rateLimitKey).(uuid.UUID); !ok {
		if config.EntryGuardEnabled() {
			http.Redirect(w, r, "/jail", http.StatusSeeOther)
		} else {
			app.SessionManager.Put(r.Context(), rateLimitKey, uuid.New())
		}
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func (app *Application) Jail(w http.ResponseWriter, r *http.Request) {
	indexes, src := jail.Get()
	app.SessionManager.Put(r.Context(), jailSolutionIndexesKey, indexes)

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Jail(tc, src).Render(r.Context(), w)
}

func (app *Application) HandleJail(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.JailForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	logger := getLoggerFromRequest(r)
	l := app.getLocalizer(ctx)

	indexes, ok := app.SessionManager.Pop(ctx, jailSolutionIndexesKey).([]int)
	if !ok {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if config.CaptchaEnabled() {
		solution := jail.GetSolution(indexes)
		if solution == "" || fd.Data.Characters != solution {
			app.notifyUser(ctx, true, l.Translate("Failed to open the gate: invalid missing characters"))
			logger.Debug("failed to solve entry guard", "solution", solution, "answer", fd.Data.Characters)
			http.Redirect(w, r, "/jail", http.StatusSeeOther)
			return
		}
	}

	app.SessionManager.Put(ctx, rateLimitKey, uuid.New())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *Application) Faq(w http.ResponseWriter, r *http.Request) {
	l := app.getLocalizer(r.Context())

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Faq(tc, support.GetFaqs(&l)).Render(r.Context(), w)
}

func (app *Application) Register(w http.ResponseWriter, r *http.Request) {
	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Register(tc).Render(r.Context(), w)
}

func (app *Application) HandleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := app.getLocalizer(ctx)

	fd, err := processForm[form.RegisterForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	err = app.DoTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := market.RegisterUser(ctx, tx, app.RiverClient, market.RegisterUserParams{
			Username: fd.Data.Username,
			Password: fd.Data.Password,
		})
		return err
	})

	if err != nil {
		if errors.Is(err, auth.ErrUsernameAlreadyRegistered) {
			app.notifyUser(ctx, true, l.Translate("This username is already in use"))
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		} else if errors.Is(err, auth.ErrInvalidPGPKey) {
			app.notifyUser(ctx, true, l.Translate("Invalid PGP key"))
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		} else {
			app.serverError(w, r, err)
			return
		}
	}

	app.notifyUser(ctx, false, l.Translate("Registered successfully"))
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Login(tc).Render(r.Context(), w)
}

func (app *Application) HandleLogin(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.LoginForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)

	var user repo.User
	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		u, err := auth.Authenticate(ctx, qtx, auth.AuthenticateParams{Username: fd.Data.Username, Password: fd.Data.Password})
		user = u
		return err
	})

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			app.notifyUser(ctx, true, l.Translate("Invalid credentials"))
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		} else if errors.Is(err, auth.ErrAccountIsBanned) {
			app.notifyUser(ctx, true, l.Translate("Your account is banned"))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else {
			app.serverError(w, r, err)
			return
		}
	}

	if user.TwofaEnabled {
		challenge, err := auth.Generate2FAChallenge(user.PgpKey)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		app.SessionManager.Put(ctx, twoFactorTokenKey, challenge.Token)
		app.SessionManager.Put(ctx, twoFactorUIDKey, user.ID)

		tc, err := app.newTemplateContext(r)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		ui.TwoFA(tc, challenge.EncryptedMessage).Render(r.Context(), w)
	} else {
		app.login(ctx, user.ID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func (app *Application) Handle2FA(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.TwoFactorAuthForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	correctToken := app.SessionManager.PopString(ctx, twoFactorTokenKey)

	if len(correctToken) != auth.TokenLength {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if fd.Data.Token == correctToken {
		uid, ok := app.SessionManager.Pop(ctx, twoFactorUIDKey).(uuid.UUID)
		if !ok {
			app.serverError(w, r, fmt.Errorf("2fa failed"))
			return
		}

		app.login(ctx, uid)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		l := app.getLocalizer(ctx)
		app.notifyUser(r.Context(), true, l.Translate("Invalid 2FA token"))
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func (app *Application) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if err := app.logout(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.ChangePasswordForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)
	user := app.loggedInUser(ctx)

	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		_, err := auth.ChangePassword(ctx, qtx, auth.ChangePasswordParams{
			Username:    user.Username,
			OldPassword: fd.Data.Password,
			NewPassword: fd.Data.NewPassword})
		return err
	})

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) || errors.Is(err, auth.ErrInvalidPassword) {
			app.notifyUser(ctx, true, l.Translate("Invalid credentials"))
			http.Redirect(w, r, "/settings", http.StatusSeeOther)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(ctx, false, l.Translate("Password changed"))
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (app *Application) HandleUpdatePGP(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.PGPForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	challenge, err := auth.Generate2FAChallenge(fd.Data.PgpKey)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ctx := r.Context()
	app.SessionManager.Put(ctx, pgpKeyKey, fd.Data.PgpKey)
	app.SessionManager.Put(ctx, twoFactorTokenKey, challenge.Token)

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.PGPCheck(tc, challenge.EncryptedMessage).Render(r.Context(), w)
}

func (app *Application) HandleCheckPGP(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.TwoFactorAuthForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	correctToken := app.SessionManager.PopString(ctx, twoFactorTokenKey)

	if len(correctToken) != auth.TokenLength {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	l := app.getLocalizer(ctx)

	if fd.Data.Token == correctToken {
		newKey := app.SessionManager.PopString(ctx, pgpKeyKey)
		if newKey == "" {
			app.serverError(w, r, fmt.Errorf("failed to get pgp key from session"))
			return
		}

		user := app.loggedInUser(ctx)
		oldKey := user.PgpKey

		if newKey == oldKey {
			app.notifyUser(r.Context(), true, l.Translate("You can't update to your current key"))
			http.Redirect(w, r, "/settings", http.StatusSeeOther)
			return
		}

		err := app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
			_, err := auth.SetPGPKey(ctx, qtx, auth.SetPGPKeyParams{
				UserID: user.ID,
				PgpKey: newKey,
			})
			return err
		})
		if err != nil {
			if errors.Is(err, auth.ErrInvalidPGPKey) {
				app.notifyUser(ctx, true, l.Translate("Invalid PGP key"))
				http.Redirect(w, r, "/settings", http.StatusSeeOther)
				return
			}
			app.serverError(w, r, err)
			return
		}

		if oldKey == "" {
			app.notifyUser(ctx, false, l.Translate("PGP key is set"))
		} else {
			app.notifyUser(ctx, false, l.Translate("PGP Public key is updated"))
		}

		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	} else {
		app.notifyUser(r.Context(), true, l.Translate("Invalid 2FA token"))
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	}
}

func (app *Application) HandleUpdateTermsOfService(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.UpdateTermsOfService](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)

	user := app.loggedInUser(ctx)

	if fd.Data.TermsOfService != "" {
		err := app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
			_, err := app.Repo().CreateTermsOfService(ctx, repo.CreateTermsOfServiceParams{
				Content:  fd.Data.TermsOfService,
				VendorID: user.ID,
			})
			return err
		})
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		app.notifyUser(ctx, false, l.Translate("Terms of service updated"))
	}

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (app *Application) ViewCreateListing(w http.ResponseWriter, r *http.Request) {
	categories, err := app.Repo().GetCategories(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	ui.CreateListing(tc, product.SupportedLocations(), categories).Render(r.Context(), w)
}

func (app *Application) HandleCreateListing(w http.ResponseWriter, r *http.Request) {
	l := app.getLocalizer(r.Context())
	ctx := r.Context()

	fd, err := processForm[form.CreateListingForm](r, app.SessionManager, form.ListingFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(ctx, app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	user := app.loggedInUser(ctx)

	priceTiers := []product.PriceTier{}
	// Convert price_tiers from users currency to cents in default currency
	for _, pt := range fd.Data.PriceTiers {
		if pt.Quantity <= 0 || pt.Price <= 0 {
			continue
		}
		priceTiers = append(priceTiers, product.PriceTier{
			PriceCent: currency.Fiat2Fiat(currency.Currency(user.Currency), currency.DefaultCurrency, pt.Price*100),
			Quantity:  pt.Quantity,
		})
	}

	product, err := product.Create(ctx, app.Repo(), app.MinioClient, product.CreateParams{
		Title:       fd.Data.Title,
		Description: fd.Data.Description,
		CategoryID:  fd.Data.CategoryID,
		Inventory:   fd.Data.Inventory,
		ShipsFrom:   fd.Data.ShipsFrom,
		ShipsTo:     fd.Data.ShipsTo,
		VendorID:    user.ID,
		Image:       fd.Files["image"].File,
		PriceTiers:  priceTiers,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(ctx, false, l.Translate("Listing created"))
	http.Redirect(w, r, fmt.Sprintf("/product?id=%s", product.ID), http.StatusSeeOther)
}

func (app *Application) Products(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	if err != nil || page < 1 {
		page = 1
	}

	ctx := r.Context()
	queries := repo.New(app.Db)

	allProducts, err := view.V.Product.GetAll(ctx, queries)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	products, numPages, err := util.GetPage(allProducts, int(page), 32)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Products(tc, products, int(page), numPages).Render(ctx, w)
}

func (app *Application) Product(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	priceTier, _ := strconv.Atoi(r.URL.Query().Get("pt"))

	ctx := r.Context()
	queries := repo.New(app.Db)

	product, err := view.V.Product.Get(ctx, queries, id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Product(tc, &product, priceTier, tc.UID == product.VendorID).Render(ctx, w)
}

func (app *Application) Cart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := repo.New(app.Db)
	customer := app.loggedInUser(ctx)

	items, err := q.GetViewCartForCustomer(ctx, customer.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	vendorGroups := map[uuid.UUID][]repo.GetViewCartForCustomerRow{}
	for _, item := range items {
		vendorGroups[item.User.ID] = append(vendorGroups[item.User.ID], item)
	}

	groups := [][]repo.GetViewCartForCustomerRow{}
	for _, group := range vendorGroups {
		groups = append(groups, group)
	}

	slices.SortFunc(groups, func(a, b []repo.GetViewCartForCustomerRow) int {
		var aTotal, bTotal int64
		for _, item := range a {
			aTotal += item.PriceCent * item.Count
		}

		for _, item := range b {
			bTotal += item.PriceCent * item.Count
		}

		return int(bTotal - aTotal)
	})

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Cart(tc, groups).Render(ctx, w)
}

func (app *Application) Checkout(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	q := repo.New(app.Db)

	dms, err := q.GetDeliveryMethodsForVendor(ctx, vendorID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Checkout(tc, dms).Render(ctx, w)
}

func (app *Application) HandleUpdateCart(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.UpdateCartForm](r, app.SessionManager, form.ListingFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(r.Context())
	q := repo.New(app.Db)

	customer := app.loggedInUser(ctx)

	if fd.Data.Action == "add" {
		_, err := q.CreateCartItem(ctx, repo.CreateCartItemParams{
			CustomerID: customer.ID,
			PriceID:    fd.Data.PriceTierID,
		})
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	} else {
		err := q.RemoveCartItem(ctx, repo.RemoveCartItemParams{
			CustomerID: customer.ID,
			PriceID:    fd.Data.PriceTierID,
		})
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	app.notifyUser(ctx, false, l.Translate("Cart updated"))
	app.redirectBack(w, r)
}

func (app *Application) HandleDeleteCart(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.IDForm](r, app.SessionManager, form.ListingFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(r.Context())
	q := repo.New(app.Db)

	customer := app.loggedInUser(ctx)

	err = q.RemoveCartItems(ctx, repo.RemoveCartItemsParams{
		CustomerID: customer.ID,
		VendorID:   fd.Data.ID,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(ctx, false, l.Translate("cart deleted"))
	app.redirectBack(w, r)
}

func (app *Application) HandleProductAction(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.ProductActionForm](r, app.SessionManager, form.ListingFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(r.Context())

	customer := app.loggedInUser(ctx)
	if vendor.HasLicense(ctx, repo.New(app.Db), customer.ID) {
		app.notifyUser(r.Context(), true, l.Translate("You can't order with a vendor account"))
		app.redirectBack(w, r)
		return
	}

	q := repo.New(app.Db)
	if fd.Data.Action == "buy_now" {
		// Redirect to checkout
		http.Redirect(w, r, "/checkout?pt="+fd.Data.PriceID.String(), http.StatusSeeOther)
	} else {
		// Add item to cart
		_, err := q.CreateCartItem(ctx, repo.CreateCartItemParams{
			CustomerID: customer.ID,
			PriceID:    fd.Data.PriceID,
		})
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		app.redirectBack(w, r)
	}
}

func (app *Application) HandleOrder(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.OrderForm](r, app.SessionManager, form.ListingFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(r.Context())

	customer := app.loggedInUser(ctx)
	if vendor.HasLicense(ctx, repo.New(app.Db), customer.ID) {
		app.notifyUser(r.Context(), true, l.Translate("You can't order with a vendor account"))
		app.redirectBack(w, r)
		return
	}

	var newOrder repo.Order
	err = app.DoTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		qtx := repo.New(tx)

		newOrder, err = market.CreateOrder(ctx, tx, app.RiverClient, market.CreateOrderParams{
			CustomerID:       customer.ID,
			DeliveryMethodID: fd.Data.DeliveryMethodID,
			Details:          fd.Data.Details,
			UseWallet:        fd.Data.UseWallet,
		})
		if err != nil {
			return err
		}

		return qtx.RemoveCartItems(ctx, repo.RemoveCartItemsParams{
			CustomerID: customer.ID,
			VendorID:   newOrder.VendorID,
		})
	})

	if err != nil {
		if errors.Is(err, order.ErrNotEnoughBalance) {
			app.notifyUser(r.Context(), true, l.Translate("You don't have enough balance for this order"))
			app.redirectBack(w, r)
			return
		} else if errors.Is(err, order.ErrCartIsEmpty) {
			app.clientError(w, http.StatusBadRequest)
			return
		} else {
			app.serverError(w, r, err)
			return
		}
	}

	if newOrder.Status == repo.OrderStatusPending {
		app.notifyUser(r.Context(), false, l.Translate("Please pay the invoice below in %d hours for your order to proceed.", config.InvoicePaymentWindow/time.Hour))
	} else {
		app.notifyUser(r.Context(), false, l.Translate("Order created"))
	}

	http.Redirect(w, r, fmt.Sprintf("/order?id=%s", newOrder.ID), http.StatusSeeOther)
}

func (app *Application) HandleOrderChat(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.OrderChatForm](r, app.SessionManager, form.ListingFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	author := app.loggedInUser(ctx)

	q := repo.New(app.Db)

	if !order.IsCustomer(ctx, q, order.IsCustomerParams{UserID: author.ID, OrderID: fd.Data.OrderID}) &&
		!order.IsVendor(ctx, q, order.IsVendorParams{UserID: author.ID, OrderID: fd.Data.OrderID}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	_, err = q.CreateOrderChatMessage(ctx, repo.CreateOrderChatMessageParams{
		Message:  fd.Data.Message,
		AuthorID: author.ID,
		OrderID:  fd.Data.OrderID,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.redirectBack(w, r)
}

func (app *Application) ViewOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	q := repo.New(app.Db)
	user := app.loggedInUser(r.Context())

	if !order.IsCustomerOrVendor(ctx, q, order.IsCustomerOrVendorParams{UserID: user.ID, OrderID: id}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	order, err := view.V.Order.Get(ctx, q, id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Order(tc, &order, false).Render(ctx, w)
}

func (app *Application) Orders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := app.loggedInUser(ctx)
	q := repo.New(app.Db)
	isVendor := vendor.HasLicense(ctx, q, user.ID)

	var orders []view.Order
	var err error
	if isVendor {
		orders, err = view.V.Order.GetAllForVendor(ctx, q, user.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

	} else {
		orders, err = view.V.Order.GetAllForCustomer(ctx, q, user.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Orders(tc, orders, isVendor).Render(ctx, w)
}

func (app *Application) Wallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := repo.New(app.Db)

	user := app.loggedInUser(ctx)

	wallet, err := q.GetWalletForUser(ctx, user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	deposit, err := q.GetDepositForWallet(ctx, wallet.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Wallet(tc, deposit.Invoice.Address).Render(ctx, w)
}

func (app *Application) HandleWithdrawal(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.WithdrawForm](r, app.SessionManager, form.ListingFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)
	user := app.loggedInUser(r.Context())

	var amount int64
	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		amount, err = payment.WithdrawFunds(ctx, qtx, user.ID, fd.Data.Address, currency.XMR2Int(fd.Data.AmountXMR))
		return err
	})

	if errors.Is(err, payment.ErrWithdrawalAmountTooSmall) {
		app.notifyUser(ctx, true, l.Translate("Withdrawal amount is too small"))
		http.Redirect(w, r, "/wallet", http.StatusSeeOther)
		return
	} else if errors.Is(err, payment.ErrNotEnoughBalanceToWithdraw) {
		app.notifyUser(ctx, true, l.Translate("Not enough balance to withdraw."))
		http.Redirect(w, r, "/wallet", http.StatusSeeOther)
		return
	} else if errors.Is(err, payment.ErrWithdrawToOwnAddress) {
		app.notifyUser(ctx, true, l.Translate("You can't withdraw to your own account"))
		http.Redirect(w, r, "/wallet", http.StatusSeeOther)
		return
	} else if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(ctx, false, l.Translate("Withdrawal of %s XMR initiated", currency.XMR2Decimal(amount)))
	http.Redirect(w, r, "/wallet", http.StatusSeeOther)
}

func (app *Application) HandleCancelOrder(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.IDForm](r, app.SessionManager, form.ListingFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)
	q := repo.New(app.Db)

	user := app.loggedInUser(ctx)
	if !order.IsCustomer(ctx, q, order.IsCustomerParams{UserID: user.ID, OrderID: fd.Data.ID}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		return order.Cancel(ctx, qtx, fd.Data.ID)
	})

	if err != nil {
		if errors.Is(err, order.ErrInvalidStatus) {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(ctx, false, l.Translate("Order cancelled"))
	app.redirectBack(w, r)
}

func (app *Application) Review(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	q := repo.New(app.Db)
	user := app.loggedInUser(r.Context())

	if !order.IsCustomer(ctx, q, order.IsCustomerParams{UserID: user.ID, OrderID: id}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Review(tc, id).Render(r.Context(), w)
}

func (app *Application) HandleReview(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.ReviewForm](r, app.SessionManager, form.ListingFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)
	q := repo.New(app.Db)

	user := app.loggedInUser(ctx)
	if !order.IsCustomer(ctx, q, order.IsCustomerParams{UserID: user.ID, OrderID: fd.Data.OrderID}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	productReviews := map[uuid.UUID]order.ProductReview{}
	for _, pr := range fd.Data.ProductReviews {
		productReviews[pr.ItemID] = order.ProductReview{
			Grade:   pr.Grade,
			Comment: pr.Comment,
		}
	}

	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		if err := order.Finalize(ctx, qtx, fd.Data.OrderID); err != nil {
			return err
		}
		return order.CreateReview(ctx, qtx, order.CreateReviewParams{
			OrderID:        fd.Data.OrderID,
			Grade:          fd.Data.Grade,
			Comment:        fd.Data.Comment,
			ProductReviews: productReviews,
		})
	})

	if err != nil {
		if errors.Is(err, order.ErrInvalidStatus) {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(r.Context(), false, l.Translate("Review created"))
	http.Redirect(w, r, fmt.Sprintf("/order?id=%s", fd.Data.OrderID), http.StatusSeeOther)
}

func (app *Application) HandleExtend(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.IDForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)
	q := repo.New(app.Db)

	user := app.loggedInUser(ctx)

	if !order.IsCustomer(ctx, q, order.IsCustomerParams{UserID: user.ID, OrderID: fd.Data.ID}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = order.Extend(ctx, q, fd.Data.ID)
	if errors.Is(err, order.ErrExtendUnavailable) {
		app.notifyUser(r.Context(), true, l.Translate("You can't extend AF timer before order has been dispatched for %d days.", config.ExtendUnavailableWindow/(time.Hour*24)))
	} else if errors.Is(err, order.ErrReExtendUnavailable) {
		app.notifyUser(r.Context(), true, l.Translate("You can't re-extend AF timer before order has been dispatched for %d days.", (config.ExtendUnavailableWindow+config.OrderDeliveryWindow)/(time.Hour*24)))
	} else if errors.Is(err, order.ErrUnableToExtendFurther) {
		app.notifyUser(r.Context(), true, l.Translate("AF timer already extended to maximum."))
	} else if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(r.Context(), false, l.Translate("AF timer extended by %d days", config.OrderDeliveryWindow/(time.Hour*24)))
	app.redirectBack(w, r)
}

func (app *Application) HandleDispute(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.IDForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)
	q := repo.New(app.Db)

	user := app.loggedInUser(ctx)
	if !order.IsCustomer(ctx, q, order.IsCustomerParams{UserID: user.ID, OrderID: fd.Data.ID}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := order.Dispute(ctx, q, fd.Data.ID); err != nil {
		if errors.Is(err, order.ErrInvalidStatus) {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(ctx, false, l.Translate("Order disputed"))
	app.redirectBack(w, r)
}

func (app *Application) HandleDisputeOffer(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.DisputeOfferForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	q := repo.New(app.Db)

	user := app.loggedInUser(r.Context())
	if !order.IsVendor(ctx, q, order.IsVendorParams{UserID: user.ID, OrderID: fd.Data.OrderID}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	_, err = order.CreateDisputeOffer(ctx, q, order.CreateDisputeOfferParams{
		OrderID:      fd.Data.OrderID,
		RefundFactor: fd.Data.RefundFactor,
	})
	if err != nil {
		if errors.Is(err, order.ErrInvalidStatus) {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.redirectBack(w, r)
}

func (app *Application) HandleDisputeOfferResponse(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.DisputeOfferResponseForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	q := repo.New(app.Db)

	offer, err := q.GetDisputeOffer(ctx, fd.Data.OfferID)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r.Context())
	if !order.IsCustomer(ctx, q, order.IsCustomerParams{UserID: user.ID, OrderID: offer.OrderID}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		if fd.Data.Accept {
			return order.AcceptDisputeOffer(ctx, qtx, offer.ID)
		} else {
			return order.DeclineDisputeOffer(ctx, qtx, offer.ID)
		}
	})
	if err != nil {
		if errors.Is(err, order.ErrInvalidStatus) {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.redirectBack(w, r)
}

func (app *Application) VendorApplication(w http.ResponseWriter, r *http.Request) {
	q := repo.New(app.Db)
	settings, err := settings.Get(r.Context(), q)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	price := currency.XMR2Decimal(settings.VendorApplicationPrice)

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Application(tc, price).Render(r.Context(), w)
}

func (app *Application) HandleVendorApplication(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.VendorApplicationForm](r, app.SessionManager, form.VendorApplicationFormFileRules)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)

	user := app.loggedInUser(ctx)

	if !user.TwofaEnabled {
		app.notifyUser(ctx, true, l.Translate("2FA needs to be enabled for vendors"))
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	queries := repo.New(app.Db)

	orders, err := queries.GetOrdersForCustomer(ctx, user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if len(orders) > 0 {
		app.notifyUser(ctx, true, l.Translate("Vendor accounts can't have previous orders. Please create a new account."))
		app.redirectBack(w, r)
		return
	}

	if !fd.Data.ExistingVendor {
		if _, ok := fd.Files["inventory"]; !ok {
			app.notifyUser(ctx, true, l.Translate("Proof of inventory is required from new vendors"))
			app.redirectBack(w, r)
			return
		}
	}

	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		var inventory io.ReadSeeker
		if !fd.Data.ExistingVendor {
			inventory = fd.Files["inventory"].File
		}
		_, err := vendor.CreateApplication(ctx, qtx, app.MinioClient, vendor.CreateApplicationParams{
			UserID:         user.ID,
			Logo:           fd.Files["logo"].File,
			InventoryImage: inventory,
			ExistingVendor: fd.Data.ExistingVendor,
			Letter:         fd.Data.Letter,
		})
		return err
	})
	if errors.Is(err, vendor.ErrNotEnoughBalance) {
		app.notifyUser(ctx, true, l.Translate("Not enough balance"))
		app.redirectBack(w, r)
		return
	} else if errors.Is(err, vendor.ErrUserHasAlreadyApplied) {
		app.notifyUser(ctx, true, l.Translate("Not enough balance"))
		app.redirectBack(w, r)
		return
	} else if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(r.Context(), false, l.Translate("Application received"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) HandleProcessOrder(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.ProcessForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)
	q := repo.New(app.Db)
	user := app.loggedInUser(ctx)

	if !order.IsVendor(ctx, q, order.IsVendorParams{UserID: user.ID, OrderID: fd.Data.ID}) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		if fd.Data.Accept {
			return order.Accept(ctx, qtx, fd.Data.ID)
		} else {
			return order.Decline(ctx, qtx, fd.Data.ID)
		}
	})
	if errors.Is(err, order.ErrInvalidStatus) {
		app.clientError(w, http.StatusBadRequest)
		return
	} else if err != nil {
		app.serverError(w, r, err)
		return
	}

	if fd.Data.Accept {
		notification.OrderAccepted(ctx, q, fd.Data.ID)
		app.notifyUser(ctx, false, l.Translate("Order accepted. You have %d hours to mark it dispatched.", config.OrderDispatchWindow/time.Hour))

	} else {
		app.notifyUser(ctx, false, l.Translate("Order declined"))
	}

	app.redirectBack(w, r)
}

func (app *Application) HandleDeliver(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.DispatchForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	q := repo.New(app.Db)
	user := app.loggedInUser(ctx)

	if !order.IsVendor(ctx, q, order.IsVendorParams{UserID: user.ID, OrderID: fd.Data.OrderID}) {

		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		return order.Dispatch(ctx, qtx, fd.Data.OrderID)
	})
	if errors.Is(err, order.ErrInvalidStatus) {
		app.clientError(w, http.StatusBadRequest)
		return
	} else if err != nil {
		app.serverError(w, r, err)
		return
	}

	notification.OrderDispatched(ctx, q, fd.Data.OrderID)

	l := app.getLocalizer(ctx)
	app.notifyUser(ctx, false, l.Translate("Order marked as dispatched"))
	app.redirectBack(w, r)
}

func (app *Application) HandleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.UpdateProductForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	q := repo.New(app.Db)

	p, err := q.GetProduct(ctx, fd.Data.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	user := app.loggedInUser(ctx)
	if user.ID != p.VendorID {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	l := app.getLocalizer(ctx)

	if fd.Data.Delete {
		err := app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
			return product.Delete(ctx, qtx, fd.Data.ID)
		})
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		app.notifyUser(ctx, false, l.Translate("Product deleted"))
		http.Redirect(w, r, "/products", http.StatusSeeOther)
	} else {
		err := app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {

			return product.UpdateInventory(ctx, qtx, p.ID, fd.Data.Inventory)
		})
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		app.notifyUser(ctx, false, l.Translate("Product updated"))
		app.redirectBack(w, r)
	}
}

func (app *Application) CreateTicket(w http.ResponseWriter, r *http.Request) {
	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.CreateTicket(tc, config.PgpKey).Render(r.Context(), w)
}

func (app *Application) HandleCreateTicket(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.TicketForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	queries := repo.New(app.Db)

	user := app.loggedInUser(r.Context())

	ticket, err := queries.CreateTicket(
		ctx,
		repo.CreateTicketParams{
			Subject:  fd.Data.Subject,
			Message:  fd.Data.Message,
			AuthorID: user.ID,
		})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/support/ticket?id="+ticket.ID.String(), http.StatusSeeOther)
}

func (app *Application) Ticket(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	ticket, err := view.V.Ticket.Get(ctx, repo.New(app.Db), id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	user := app.loggedInUser(r.Context())
	if user.ID != ticket.Ticket.AuthorID {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Ticket(tc, ticket).Render(ctx, w)
}

func (app *Application) Tickets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := repo.New(app.Db)
	author := app.loggedInUser(ctx)
	tickets, err := queries.GetTicketsForAuthor(ctx, author.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Tickets(tc, tickets).Render(ctx, w)
}

func (app *Application) HandleTicketResponse(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.TicketResponseForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	l := app.getLocalizer(ctx)
	queries := repo.New(app.Db)

	user := app.loggedInUser(r.Context())
	ticket, err := queries.GetTicket(ctx, fd.Data.TicketID)

	if err != nil {
		app.serverError(w, r, err)
		return
	} else if user.ID != ticket.AuthorID {
		app.clientError(w, http.StatusBadRequest)
		return
	} else if !ticket.IsOpen {
		app.notifyUser(r.Context(), true, l.Translate("Ticket is already closed"))
		app.redirectBack(w, r)
		return
	}

	_, err = queries.CreateTicketResponse(
		ctx,
		repo.CreateTicketResponseParams{
			Message:    fd.Data.Message,
			TicketID:   fd.Data.TicketID,
			AuthorName: user.Username,
		})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.redirectBack(w, r)
}

func (app *Application) Settings(w http.ResponseWriter, r *http.Request) {
	currencies := currency.SupportedCurrencies()
	deliveryMethods := []repo.DeliveryMethod{}

	ctx := r.Context()
	q := repo.New(app.Db)
	user := app.loggedInUser(ctx)
	if vendor.HasLicense(ctx, q, user.ID) {
		var err error
		deliveryMethods, err = q.GetDeliveryMethodsForVendor(ctx, user.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Settings(tc, currencies, deliveryMethods).Render(ctx, w)
}

func (app *Application) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.SettingsForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	user := app.loggedInUser(ctx)

	if fd.Data.Locale != user.Locale || fd.Data.Currency != user.Currency || fd.Data.Enable2FA != user.TwofaEnabled {
		err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
			return user_settings.UpdateSettings(ctx, qtx, user_settings.UpdateSettingsParams{
				ID:               user.ID,
				Locale:           fd.Data.Locale,
				Currency:         fd.Data.Currency,
				TwofaEnabled:     fd.Data.Enable2FA,
				IncognitoEnabled: fd.Data.EnableIncognito,
			})
		})
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		app.SessionManager.Put(ctx, localeKey, fd.Data.Locale)
	}

	app.redirectBack(w, r)
}

func (app *Application) HandleUpdateDeliveryMethods(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.DeliveryMethodsForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	user := app.loggedInUser(ctx)

	dms := []vendor.DeliveryMethod{}
	for _, dm := range fd.Data.DeliveryMethods {
		if dm.Description == "" {
			continue
		}
		priceCent := currency.Fiat2Fiat(currency.Currency(user.Currency), currency.DefaultCurrency, int64(dm.Price*100))
		dms = append(dms, vendor.DeliveryMethod{
			Description: dm.Description,
			PriceCent:   priceCent,
		})
	}

	err = app.Do(ctx, func(ctx context.Context, qtx *repo.Queries) error {
		return vendor.UpdateDeliveryMethods(ctx, qtx, vendor.UpdateDeliveryMethodsParams{
			VendorID:        user.ID,
			DeliveryMethods: dms,
		})
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (app *Application) Vendor(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	vendor, err := view.V.Vendor.Get(ctx, repo.New(app.Db), id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Vendor(tc, &vendor).Render(ctx, w)
}

func (app *Application) Notifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := app.loggedInUser(ctx)

	q := repo.New(app.Db)

	notifications, err := q.GetNotificationsForUser(ctx, user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Notifications(tc, notifications).Render(ctx, w)
}

func (app *Application) HandleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.IDForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer fd.Close()

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		app.redirectBack(w, r)
		return
	}

	ctx := r.Context()
	q := repo.New(app.Db)
	if err := q.DeleteNotification(ctx, fd.Data.ID); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func (app *Application) HandleDeleteAllNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := app.loggedInUser(r.Context())
	q := repo.New(app.Db)

	if err := q.DeleteAllNotificationsForUser(ctx, user.ID); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func (app *Application) Health(w http.ResponseWriter, r *http.Request) {
	health := model.Health{Status: http.StatusOK}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := app.Db.Ping(ctx); err != nil {
		health.Status = http.StatusServiceUnavailable
		health.Service.Database = false
	} else {
		health.Service.Database = true
	}

	if err := json.NewEncoder(w).Encode(&health); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *Application) ServeUpload(w http.ResponseWriter, r *http.Request) {
	p := path.Clean(r.URL.Path)

	//TODO: Check if p is in a cache

	upload, err := util.LoadUpload(r.Context(), app.MinioClient, path.Dir(p), path.Base(p))
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	defer upload.Close()

	http.ServeContent(w, r, p, time.Time{}, upload)
}
