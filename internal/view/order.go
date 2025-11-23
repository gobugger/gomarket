package view

import (
	"context"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/repo"
	"slices"
	"time"
)

type OrderChatItem struct {
	Message    *repo.OrderChatMessage
	Offer      *repo.DisputeOffer
	AuthorName string
}

type OrderItem struct {
	Product   repo.Product
	PriceCent int64
	Quantity  int32
	Count     int32
}

type Order struct {
	Order             repo.Order
	Items             []OrderItem
	DeliveryMethod    repo.DeliveryMethod
	Customer          repo.User
	Vendor            repo.User
	Invoice           *Invoice
	TimeUntilFinalize time.Duration
	TimeUntilDecline  time.Duration
	CanExtend         bool
	ChatItems         []OrderChatItem
}

type OrderView struct{}

func isDispatched(status repo.OrderStatus) bool {
	return status == repo.OrderStatusDispatched ||
		status == repo.OrderStatusFinalized ||
		status == repo.OrderStatusDisputed ||
		status == repo.OrderStatusSettled
}

func extendView(ctx context.Context, q *repo.Queries, view *Order) error {
	status := view.Order.Status

	switch status {
	case repo.OrderStatusPending:
		invoice, err := InvoiceView{}.GetWithOrderID(ctx, q, view.Order.ID)
		if err != nil {
			return err
		}
		view.Invoice = &invoice
	case repo.OrderStatusPaid:
		view.TimeUntilDecline = config.OrderProcessingWindow - time.Since(view.Order.CreatedAt)
	case repo.OrderStatusAccepted:
		view.TimeUntilDecline = config.OrderProcessingWindow - time.Since(view.Order.AcceptedAt)
	}

	items, err := q.GetOrderItems(ctx, view.Order.ID)
	if err != nil {
		return err
	}

	for _, item := range items {
		product, err := q.GetProduct(ctx, item.ProductID)
		if err != nil {
			return err
		}

		view.Items = append(view.Items, OrderItem{
			Product:   product,
			PriceCent: item.PriceCent,
			Quantity:  item.Quantity,
			Count:     item.Count,
		})
	}

	if isDispatched(status) {
		extension := time.Duration(view.Order.NumExtends) * config.OrderDeliveryWindow
		view.TimeUntilFinalize = config.OrderDeliveryWindow + extension - time.Since(view.Order.DispatchedAt)
		view.CanExtend = view.TimeUntilFinalize < config.OrderDeliveryWindow-config.ExtendUnavailableWindow
	}

	return nil
}

func extendChat(ctx context.Context, q *repo.Queries, view *Order) error {
	chatItems := []OrderChatItem{}

	orderID := view.Order.ID
	{
		rows, err := q.GetOrderChatMessagesForOrderJoinAuthor(ctx, orderID)
		if err != nil {
			return err
		}

		for _, row := range rows {
			chatItems = append(chatItems, OrderChatItem{
				Message:    &row.OrderChatMessage,
				AuthorName: row.User.Username,
			})
		}
	}

	{
		offers, err := q.GetDisputeOffersForOrder(ctx, orderID)
		if err != nil {
			return err
		}

		authorName := ""

		if len(offers) > 0 {
			vendor, err := q.GetVendorForOrder(ctx, orderID)
			if err != nil {
				return err
			}
			authorName = vendor.Username
		}

		for _, row := range offers {
			chatItems = append(chatItems, OrderChatItem{
				Offer:      &row,
				AuthorName: authorName,
			})
		}
	}

	// Sort items by createdAt
	slices.SortFunc(chatItems, func(a OrderChatItem, b OrderChatItem) int {
		createdAt := func(item *OrderChatItem) time.Time {
			if item.Message != nil {
				return item.Message.CreatedAt
			} else if item.Offer != nil {
				return item.Offer.CreatedAt
			} else {
				return time.Time{}
			}
		}

		ac := createdAt(&a)
		bc := createdAt(&b)

		if ac.After(bc) {
			return 1
		} else if ac.Equal(bc) {
			return 0
		} else {
			return -1
		}
	})

	view.ChatItems = chatItems

	return nil
}

func (ov OrderView) Get(ctx context.Context, q *repo.Queries, orderID uuid.UUID) (Order, error) {
	row, err := q.GetViewOrder(ctx, orderID)
	if err != nil {
		return Order{}, err
	}

	view := Order{
		Order:          row.Order,
		DeliveryMethod: row.DeliveryMethod,
		Customer:       row.User,
		Vendor:         row.User_2,
	}

	if err := extendView(ctx, q, &view); err != nil {
		return Order{}, err
	}

	if err := extendChat(ctx, q, &view); err != nil {
		return Order{}, err
	}

	return view, nil
}

func (ov OrderView) GetAllForCustomer(ctx context.Context, q *repo.Queries, customerID uuid.UUID) ([]Order, error) {
	os, err := q.GetOrdersForCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}

	res := []Order{}

	for _, o := range os {
		order, err := ov.Get(ctx, q, o.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, order)
	}

	slices.SortFunc(res, func(a, b Order) int {
		return int(b.Order.CreatedAt.Sub(a.Order.CreatedAt))
	})

	return res, nil
}

func (ov OrderView) GetAllForVendor(ctx context.Context, q *repo.Queries, vendorID uuid.UUID) ([]Order, error) {
	os, err := q.GetOrdersForVendor(ctx, vendorID)
	if err != nil {
		return nil, err
	}

	res := []Order{}

	for _, o := range os {
		order, err := ov.Get(ctx, q, o.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, order)
	}

	slices.SortFunc(res, func(a, b Order) int {
		return int(b.Order.CreatedAt.Sub(a.Order.CreatedAt))
	})

	return res, nil
}
