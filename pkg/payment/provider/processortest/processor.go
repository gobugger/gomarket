package processortest

import (
	"crypto/rand"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"math/big"
)

type PaymentProvider struct {
	InvoiceStatuses    map[string]*provider.InvoiceStatus
	TransferStatuses   map[string]*provider.TransferStatus
	InvoiceError       error
	DeleteInvoiceError error
	TransferError      error
}

func NewPaymentProvider() *PaymentProvider {
	return &PaymentProvider{
		InvoiceStatuses:  map[string]*provider.InvoiceStatus{},
		TransferStatuses: map[string]*provider.TransferStatus{},
	}
}

func (p *PaymentProvider) Invoice(amount *big.Int, callbackUrl string) (string, error) {
	if p.InvoiceError != nil {
		return "", p.InvoiceError
	}
	return testutil.XMRAddress(), nil
}

func (p *PaymentProvider) InvoiceStatus(address string) (*provider.InvoiceStatus, error) {
	if status, ok := p.InvoiceStatuses[address]; ok {
		return status, nil
	} else {
		return &provider.InvoiceStatus{}, nil
	}
}

func (p *PaymentProvider) DeleteInvoice(address string) error {
	return p.DeleteInvoiceError
}

func (p *PaymentProvider) Transfer(destinations []provider.Destination) (*provider.TransferResponse, error) {
	if p.TransferError != nil {
		return nil, p.TransferError
	}
	return &provider.TransferResponse{
		TxHashList: []string{rand.Text()},
	}, nil
}

func (p *PaymentProvider) TransferStatus(txHash string) (*provider.TransferStatus, error) {
	if status, ok := p.TransferStatuses[txHash]; ok {
		return status, nil
	} else {
		return &provider.TransferStatus{}, nil
	}
}
