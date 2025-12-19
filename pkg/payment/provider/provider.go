package provider

import (
	"math/big"
)

type InvoiceStatus struct {
	AmountUnlocked *big.Int
	AmountTotal    *big.Int
}

type Destination struct {
	Amount  *big.Int
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
	Invoice(amount *big.Int, callbackUrl string) (string, error)
	InvoiceStatus(address string) (*InvoiceStatus, error)
	DeleteInvoice(address string) error
	Transfer(destinations []Destination) (*TransferResponse, error)
	TransferStatus(txHash string) (*TransferStatus, error)
}
