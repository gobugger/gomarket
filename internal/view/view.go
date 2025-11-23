package view

type Views struct {
	Product           ProductView
	Order             OrderView
	Invoice           InvoiceView
	Review            ReviewView
	Vendor            VendorView
	Ticket            TicketView
	VendorApplication VendorApplicationView
}

var V Views
