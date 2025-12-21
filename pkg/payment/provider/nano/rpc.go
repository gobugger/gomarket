package nano

import (
	"bytes"
	"encoding/json"
	"github.com/shopspring/decimal"
	"net/http"
)

func rpc[T any](client *http.Client, url string, req any) (*T, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&req); err != nil {
		return nil, err
	}

	rawResp, err := client.Post(url, "application/json", &buf)
	if err != nil {
		return nil, err
	}
	defer rawResp.Body.Close()

	resp := new(T)
	if err := json.NewDecoder(rawResp.Body).Decode(resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type action struct {
	Action string `json:"action"`
}

type accountCreateRequest struct {
	action
	Wallet string `json:"wallet"`
}

type accountCreateResponse struct {
	Account string `json:"account"`
}

type accountBalanceRequest struct {
	action
	Account string `json:"account"`
}

type accountBalanceResponse struct {
	Balance    decimal.Decimal `json:"balance"`
	Pending    decimal.Decimal `json:"pending"`
	Receivable decimal.Decimal `json:"receivable"`
}

type accountRemoveRequest struct {
	action
	Wallet  string `json:"wallet"`
	Account string `json:"account"`
}

type accountRemoveResponse struct {
	Removed int `json:"removed"`
}

type sendRequest struct {
	action
	Wallet      string          `json:"wallet"`
	Source      string          `json:"source"`
	Destination string          `json:"destination"`
	Amount      decimal.Decimal `json:"amount"`
	ID          string          `json:"id"`
}

type sendResponse struct {
	Block string `json:"block"`
}

type blockInfoRequest struct {
	action
	JsonBlock bool   `json:"json_block"`
	Hash      string `json:"hash"`
}

type blockInfoResponse struct {
	action
	Amount    decimal.Decimal `json:"amount"`
	Confirmed bool            `json:"confirmed"`
	Contents  struct {
		LinkAsAccount string `json:"link_as_account"`
	} `json:"contents"`
}
