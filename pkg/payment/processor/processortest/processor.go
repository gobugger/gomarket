package processortest

import (
	"crypto/rand"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/gobugger/gomarket/pkg/payment/processor"
)

type Processor struct {
	InvoiceStatuses    map[string]*processor.InvoiceStatus
	TransferStatuses   map[string]*processor.TransferStatus
	InvoiceError       error
	DeleteInvoiceError error
	TransferError      error
}

func NewProcessor() *Processor {
	return &Processor{
		InvoiceStatuses:  map[string]*processor.InvoiceStatus{},
		TransferStatuses: map[string]*processor.TransferStatus{},
	}
}

func (p *Processor) Invoice(amount int64, callbackUrl string) (string, error) {
	if p.InvoiceError != nil {
		return "", p.InvoiceError
	}
	return testutil.XMRAddress(), nil
}

func (p *Processor) InvoiceStatus(address string) (*processor.InvoiceStatus, error) {
	if status, ok := p.InvoiceStatuses[address]; ok {
		return status, nil
	} else {
		return &processor.InvoiceStatus{}, nil
	}
}

func (p *Processor) DeleteInvoice(address string) error {
	return p.DeleteInvoiceError
}

func (p *Processor) Transfer(destinations []processor.Destination) (*processor.TransferResponse, error) {
	if p.TransferError != nil {
		return nil, p.TransferError
	}
	return &processor.TransferResponse{
		TxHashList: []string{rand.Text()},
	}, nil
}

func (p *Processor) TransferStatus(txHash string) (*processor.TransferStatus, error) {
	if status, ok := p.TransferStatuses[txHash]; ok {
		return status, nil
	} else {
		return &processor.TransferStatus{}, nil
	}
}
