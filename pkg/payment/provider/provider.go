package provider

import (
	"github.com/shopspring/decimal"
)

type InvoiceStatus struct {
	AmountUnlocked decimal.Decimal
	AmountTotal    decimal.Decimal
}

type Destination struct {
	Amount  decimal.Decimal
	Address string
}

type TransferResponse struct {
	TxHashList []string
}

type TransferStatus struct {
	Destinations  []Destination
	Confirmations uint64
	Failed        bool
}

type PaymentProvider interface {
	Invoice(amount decimal.Decimal, callbackUrl string) (string, error)
	InvoiceStatus(address string) (*InvoiceStatus, error)
	DeleteInvoice(address string) error
	Transfer(destinations []Destination) (*TransferResponse, error)
	TransferStatus(txHash string) (*TransferStatus, error)
}
