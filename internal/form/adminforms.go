package form

import (
	"encoding/gob"
	"github.com/google/uuid"
)

type AdminDisputeForm struct {
	OrderID      uuid.UUID `schema:"order_id,required"`
	RefundFactor float64   `schema:"refund_factor"`
}

type AdminTicketResponseForm struct {
	TicketID    uuid.UUID `schema:"ticket_id,required"`
	Message     string    `schema:"message,required"`
	CloseTicket bool      `schema:"close_ticket,default:false"`
}

type AdminVendorApplicationForm struct {
	ApplicationID uuid.UUID `schema:"application_id,required"`
	Accept        bool      `schema:"accept,default:false"`
	Explanation   string    `schema:"explanation"`
}

type AdminDeleteForm struct {
	Operation string    `schema:"operation,required"`
	ID        uuid.UUID `schema:"id"`
}

type AdminSettingsForm struct {
	VendorApplicationPrice int64 `schema:"vendor_application_price,required"`
}

type AdminAddCategoryForm struct {
	Name      string    `schema:"name,required"`
	AddParent bool      `schema:"add_parent"`
	ParentID  uuid.UUID `schema:"parent_id"`
}

func init() {
	gob.Register(AdminDisputeForm{})
	gob.Register(AdminTicketResponseForm{})
	gob.Register(AdminVendorApplicationForm{})
	gob.Register(AdminDeleteForm{})
	gob.Register(AdminSettingsForm{})
}
