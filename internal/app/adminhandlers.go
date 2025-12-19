package app

import (
	"github.com/gobugger/gomarket/internal/form"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/settings"
	"github.com/gobugger/gomarket/internal/service/vendor"
	"github.com/gobugger/gomarket/internal/view"
	"github.com/gobugger/gomarket/ui/templ"
	"log/slog"
	"net/http"
)

func (app *Application) Admin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := repo.New(app.Db)

	applications, err := queries.GetVendorApplications(ctx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tickets, err := queries.GetOpenTickets(ctx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	settings, err := settings.Get(ctx, queries)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	disputedOrders, err := queries.GetOrdersWithStatus(ctx, repo.OrderStatusDisputed)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	categories, err := queries.GetCategories(ctx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.AdminPanel(tc, settings.VendorApplicationPrice, applications, tickets, disputedOrders, categories).Render(ctx, w)
}

func (app *Application) AdminHandleOperation(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.AdminDeleteForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		http.Redirect(w, r, "/", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := repo.New(app.Db)

	switch fd.Data.Operation {
	case "banUser":
		if _, err := queries.CreateBan(ctx, fd.Data.ID); err != nil {
			app.serverError(w, r, err)
			return
		}
	case "deleteBan":
		if err := queries.DeleteBan(ctx, fd.Data.ID); err != nil {
			app.serverError(w, r, err)
			return
		}
	case "deleteListing":
		if err := queries.DeleteProduct(ctx, fd.Data.ID); err != nil {
			app.serverError(w, r, err)
			return
		}
	case "deleteReview":
		if err := queries.DeleteReview(ctx, fd.Data.ID); err != nil {
			app.serverError(w, r, err)
			return
		}
	default:
		slog.Error("Unknown operation", "operation", fd.Data.Operation)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) AdminDispute(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := repo.New(app.Db)

	order, err := view.V.Order.Get(ctx, queries, id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.Order(tc, &order, true).Render(ctx, w)
}

func (app *Application) AdminHandleDispute(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.AdminDisputeForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		http.Redirect(w, r, "/", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) AdminTicket(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		slog.Debug("unable to parse id from query url", slog.Any("error", err))
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	ticket, err := view.V.Ticket.Get(ctx, repo.New(app.Db), id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.HandleTicket(tc, ticket).Render(ctx, w)
}

func (app *Application) AdminHandleTicket(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.AdminTicketResponseForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		http.Redirect(w, r, "/", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	tx, err := app.Db.Begin(ctx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	qtx := repo.New(tx)

	if _, err := qtx.CreateTicketResponse(
		ctx,
		repo.CreateTicketResponseParams{
			Message:    fd.Data.Message,
			TicketID:   fd.Data.TicketID,
			AuthorName: "admin",
		}); err != nil {
		app.serverError(w, r, err)
		return
	}

	if fd.Data.CloseTicket {
		if _, err := qtx.CloseTicket(ctx, fd.Data.TicketID); err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.redirectBack(w, r)
}

func (app *Application) AdminVendorApplication(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		slog.Debug("unable to parse id from query url", slog.Any("error", err))
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	application, err := view.V.VendorApplication.Get(ctx, repo.New(app.Db), id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	tc, err := app.newTemplateContext(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ui.HandleApplication(tc, application).Render(ctx, w)

}

func (app *Application) AdminHandleVendorApplication(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.AdminVendorApplicationForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		http.Redirect(w, r, "/", http.StatusBadRequest)
		return
	}

	q := repo.New(app.Db)
	ctx := r.Context()

	defer func() {
	}()

	if fd.Data.Accept {
		l, err := vendor.AcceptApplication(r.Context(), q, fd.Data.ApplicationID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		content := fd.Data.Explanation
		if content == "" {
			content = "Congratulations, your vendor application has been accepted!"
		}

		_, err = q.CreateNotification(ctx, repo.CreateNotificationParams{
			Title:   "Vendor application",
			Content: content,
			UserID:  l.UserID,
		})
		if err != nil {
			slog.Error("Failed to create notification", slog.Any("error", err))
		}
	} else {
		application, err := q.GetVendorApplication(ctx, fd.Data.ApplicationID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		content := fd.Data.Explanation
		if content == "" {
			content = "Your vendor application has been declined."
		}

		if err := vendor.DeclineApplication(ctx, q, fd.Data.ApplicationID); err != nil {
			app.serverError(w, r, err)
			return
		}

		_, err = q.CreateNotification(ctx, repo.CreateNotificationParams{
			Title:   "Vendor application",
			Content: content,
			UserID:  application.UserID,
		})
		if err != nil {
			slog.Error("Failed to create notification", slog.Any("error", err))
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) AdminHandleSettings(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.AdminSettingsForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		http.Redirect(w, r, "/", http.StatusBadRequest)
		return
	}

	if fd.Data.VendorApplicationPrice.Sign() < 0 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	q := repo.New(app.Db)
	err = settings.Set(r.Context(), q, settings.Settings{VendorApplicationPrice: fd.Data.VendorApplicationPrice})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) AdminHandleAddCategory(w http.ResponseWriter, r *http.Request) {
	fd, err := processForm[form.AdminAddCategoryForm](r, app.SessionManager, nil)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if fd.HasErrors() {
		putForm(r.Context(), app.SessionManager, fd)
		http.Redirect(w, r, "/", http.StatusBadRequest)
		return
	}

	q := repo.New(app.Db)

	params := repo.CreateCategoryParams{Name: fd.Data.Name}
	if fd.Data.AddParent {
		params.ParentID.Scan(fd.Data.ParentID.String())
	}

	if _, err := q.CreateCategory(r.Context(), params); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
