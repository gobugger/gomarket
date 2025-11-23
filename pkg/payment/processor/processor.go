package processor

type InvoiceStatus struct {
	AmountUnlocked int64
	AmountTotal    int64
}

type Destination struct {
	Amount  int64
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

type Processor interface {
	Invoice(amount int64, callbackUrl string) (string, error)
	InvoiceStatus(address string) (*InvoiceStatus, error)
	DeleteInvoice(address string) error
	Transfer(destinations []Destination) (*TransferResponse, error)
	TransferStatus(txHash string) (*TransferStatus, error)
}
