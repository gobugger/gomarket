package moneropay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"github.com/shopspring/decimal"
	"gitlab.com/moneropay/go-monero/walletrpc"
	moneropay "gitlab.com/moneropay/moneropay/v2/pkg/model"
	"net/http"
)

type MoneropayClient struct {
	url    string
	client *http.Client
}

func NewMoneropayClient(url string) *MoneropayClient {
	return &MoneropayClient{
		url:    url,
		client: http.DefaultClient,
	}
}

func (mp *MoneropayClient) Invoice(amount decimal.Decimal, callbackUrl string) (string, error) {
	if amount.Sign() < 0 {
		return "", fmt.Errorf("can't create invoice with negative amount")
	}

	req := moneropay.ReceivePostRequest{
		Amount:      amount.BigInt().Uint64(),
		Description: "invoice",
		CallbackUrl: callbackUrl,
	}

	reqBytes, err := json.Marshal(&req)
	if err != nil {
		return "", err
	}

	url := mp.url + "/receive"
	resp, err := mp.client.Post(url, "application/json", bytes.NewReader(reqBytes)) //nolint
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid response: %v", resp)
	}

	invoice := &moneropay.ReceivePostResponse{}

	if err := json.NewDecoder(resp.Body).Decode(invoice); err != nil {
		return "", err
	}

	return invoice.Address, nil
}

func (mp *MoneropayClient) InvoiceStatus(address string) (*provider.InvoiceStatus, error) {
	url := mp.url + "/receive/" + address

	resp, err := mp.client.Get(url) //nolint
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response: %v", resp)
	}

	data := &moneropay.ReceiveGetResponse{}
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}

	//nolint
	return &provider.InvoiceStatus{
		AmountUnlocked: decimal.NewFromUint64(data.Amount.Covered.Unlocked),
		AmountTotal:    decimal.NewFromUint64(data.Amount.Covered.Total),
	}, nil

}

func (mp *MoneropayClient) DeleteInvoice(address string) error {
	url := mp.url + "/receive/" + address

	req, err := http.NewRequest(http.MethodDelete, url, nil) //nolint
	if err != nil {
		return err
	}

	resp, err := mp.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response: %v", resp)
	}

	return nil
}

func (mp *MoneropayClient) Transfer(destinations []provider.Destination) (*provider.TransferResponse, error) {
	req := moneropay.TransferPostRequest{}
	for _, dest := range destinations {
		if dest.Amount.Sign() < 0 {
			return nil, fmt.Errorf("can't transfer negative amount")
		}
		req.Destinations = append(req.Destinations,
			walletrpc.Destination{
				Amount:  dest.Amount.BigInt().Uint64(),
				Address: dest.Address,
			})
	}

	reqBytes, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}

	url := mp.url + "/transfer"
	resp, err := mp.client.Post(url, "application/json", bytes.NewReader(reqBytes)) //nolint
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response: %v", resp)
	}

	data := &moneropay.TransferPostResponse{}
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}

	return &provider.TransferResponse{
		TxHashList: data.TxHashList,
	}, nil
}

func (mp *MoneropayClient) TransferStatus(txHash string) (*provider.TransferStatus, error) {
	url := mp.url + "/transfer/" + txHash

	resp, err := mp.client.Get(url) //nolint
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response: %v", resp)
	}

	data := &moneropay.TransferGetResponse{}
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}

	return &provider.TransferStatus{
		Confirmations: data.Confirmations,
		Failed:        data.State == "failed",
	}, nil
}
