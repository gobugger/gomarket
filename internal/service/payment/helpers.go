package payment

import (
	"github.com/gobugger/gomarket/internal/config"
	"github.com/google/uuid"
)

const InvoiceRoute = "/invoice"

func MoneropayInvoiceCallbackURL(id uuid.UUID) string {
	return "http://" + config.MoneropayCallbackURL + InvoiceRoute + "?id=" + id.String()
}

func QrcodeFilename(id uuid.UUID) string {
	return id.String() + ".png"
}

type CalculatePriceParams struct {
	PriceID          uuid.UUID
	DeliveryMethodID uuid.UUID
}
