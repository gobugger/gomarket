package payment

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"github.com/gobugger/gomarket/pkg/rand"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"unicode"

	moneropay "gitlab.com/moneropay/moneropay/v2/pkg/model"
)

type FakeClient struct {
	invoices map[string]*provider.InvoiceStatus
	txes     map[string]*provider.TransferStatus
	mtx      sync.Mutex
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		invoices: map[string]*provider.InvoiceStatus{},
	}
}

func (fc *FakeClient) Invoice(amount int64, callbackUrl string) (string, error) {
	addr := fakeMoneroAddress()

	go func() {
		time.Sleep(time.Minute)

		fc.mtx.Lock()
		fc.invoices[addr] = &provider.InvoiceStatus{
			AmountTotal:    amount,
			AmountUnlocked: 0,
		}
		fc.mtx.Unlock()

		if _, err := sendCallback(amount, callbackUrl, false); err != nil {
			slog.Error("send callback", slog.Any("error", err))
		}

		time.Sleep(time.Minute)

		fc.mtx.Lock()
		fc.invoices[addr] = &provider.InvoiceStatus{
			AmountTotal:    amount,
			AmountUnlocked: amount,
		}
		fc.mtx.Unlock()

		if _, err := sendCallback(amount, callbackUrl, false); err != nil {
			slog.Error("send callback", slog.Any("error", err), "amount", amount, "url", callbackUrl)
		}
	}()

	return addr, nil
}

func (fc *FakeClient) InvoiceStatus(addr string) (*provider.InvoiceStatus, error) {
	fc.mtx.Lock()
	status, ok := fc.invoices[addr]
	fc.mtx.Unlock()

	if ok {
		return status, nil
	} else {
		return nil, fmt.Errorf("no invoice with address %s", addr)
	}
}

func (fc *FakeClient) DeleteInvoice(addr string) error {
	return nil
}

func (fc *FakeClient) Transfer(destinations []provider.Destination) (*provider.TransferResponse, error) {
	txHash := cryptorand.Text()

	fc.mtx.Lock()
	fc.txes[txHash] = &provider.TransferStatus{}
	fc.mtx.Unlock()

	go func() {
		for range 11 {
			time.Sleep(time.Second * 5)
			fc.mtx.Lock()
			fc.txes[txHash].Confirmations++
			fc.mtx.Unlock()
		}
	}()

	resp := &provider.TransferResponse{
		TxHashList: []string{txHash},
	}
	return resp, nil
}

func (fc *FakeClient) TransferStatus(txHash string) (*provider.TransferStatus, error) {
	fc.mtx.Lock()
	status, ok := fc.txes[txHash]
	fc.mtx.Unlock()

	if ok {
		return status, nil
	} else {
		return nil, fmt.Errorf("no transaction with txHash %s", txHash)
	}
}

func fakeMoneroAddress() string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
	n := byte(len(letters))

	src := make([]byte, 95)
	if _, err := cryptorand.Read(src); err != nil {
		panic(err)
	}
	for i := range src {
		src[i] = letters[src[i]%n]
	}

	runes := []rune(string(src))

	for range 10 {
		i := rand.Intn(95)
		r := runes[i]
		if unicode.IsLetter(r) && unicode.IsUpper(r) {
			runes[i] = unicode.ToLower(r)
		}
	}
	return string(runes)
}

func sendCallback(amount int64, url string, complete bool) (*http.Response, error) {

	payload := moneropay.CallbackResponse{Complete: complete}
	payload.Amount.Expected = uint64(amount)      //nolint
	payload.Amount.Covered.Total = uint64(amount) //nolint
	if complete {
		payload.Amount.Covered.Unlocked = payload.Amount.Covered.Total
	} else {
		payload.Amount.Covered.Unlocked = 0
	}

	body, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}

	return http.Post(url, "application/json", bytes.NewReader(body)) //nolint
}
