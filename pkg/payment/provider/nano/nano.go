package nano

import (
	"fmt"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"github.com/shopspring/decimal"
	"net/http"
)

type NanoClient struct {
	url     string
	wallet  string
	account string
	client  *http.Client
}

func NewNanoClient(nodeURL, wallet, account string) *NanoClient {
	return &NanoClient{
		url:     nodeURL,
		wallet:  wallet,
		account: account,
		client:  &http.Client{},
	}
}

func (nc *NanoClient) Invoice(amount decimal.Decimal, callbackUrl string) (string, error) {
	resp, err := rpc[accountCreateResponse](nc.client, nc.url, accountCreateRequest{
		action: action{"account_create"},
		Wallet: nc.wallet,
	})
	return resp.Account, err
}

func (nc *NanoClient) InvoiceStatus(address string) (*provider.InvoiceStatus, error) {
	resp, err := rpc[accountBalanceResponse](nc.client, nc.url, accountBalanceRequest{
		action:  action{"account_create"},
		Account: address,
	})

	return &provider.InvoiceStatus{
		AmountUnlocked: resp.Balance,
		AmountTotal:    resp.Balance.Add(resp.Pending).Add(resp.Receivable),
	}, err

}

func (nc *NanoClient) DeleteInvoice(address string) error {
	balanceResp, err := rpc[accountBalanceResponse](nc.client, nc.url, accountBalanceRequest{
		action:  action{"account_create"},
		Account: address,
	})
	if err != nil {
		return err
	}

	// We should not remove this account if there's funds
	if balanceResp.Pending.Sign() > 0 || balanceResp.Receivable.Sign() > 0 {
		return fmt.Errorf("Balance still pending or to be received")
	}

	if balance := balanceResp.Balance; balance.Sign() > 0 {
		_, err := rpc[sendResponse](nc.client, nc.url, sendRequest{
			action:      action{"send"},
			Wallet:      nc.wallet,
			Source:      address,
			Destination: nc.account,
			Amount:      balance,
			ID:          "",
		})
		if err != nil {
			return err
		}
	}

	_, err = rpc[accountRemoveResponse](nc.client, nc.url, accountRemoveRequest{
		action:  action{"account_remove"},
		Wallet:  nc.wallet,
		Account: address,
	})

	return err
}

func (nc *NanoClient) Transfer(destinations []provider.Destination) (*provider.TransferResponse, error) {
	result := &provider.TransferResponse{}
	for _, dst := range destinations {
		resp, err := rpc[sendResponse](nc.client, nc.url, sendRequest{
			action:      action{"send"},
			Wallet:      nc.wallet,
			Source:      nc.account,
			Destination: dst.Address,
			Amount:      dst.Amount,
			ID:          "",
		})
		if err != nil {
			return result, err
		}

		result.TxHashList = append(result.TxHashList, resp.Block)
	}

	return result, nil
}

func (nc *NanoClient) TransferStatus(txHash string) (*provider.TransferStatus, error) {
	resp, err := rpc[blockInfoResponse](nc.client, nc.url, blockInfoRequest{
		action:    action{"account_remove"},
		JsonBlock: true,
		Hash:      txHash,
	})
	if err != nil {
		return nil, err
	}

	return &provider.TransferStatus{
		Destinations: []provider.Destination{
			{
				Amount:  resp.Amount,
				Address: resp.Contents.LinkAsAccount,
			},
		},
		Confirmations: func() uint64 {
			if resp.Confirmed {
				return 10
			} else {
				return 0
			}
		}(),
	}, nil
}
