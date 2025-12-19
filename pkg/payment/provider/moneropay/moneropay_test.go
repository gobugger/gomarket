package moneropay

import (
	"encoding/json"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	moneropay "gitlab.com/moneropay/moneropay/v2/pkg/model"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMoneropayClient(t *testing.T) {
	r := http.NewServeMux()
	r.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&moneropay.ReceivePostRequest{}); err != nil {
				t.Fatal(err)
			}
			if err := json.NewEncoder(w).Encode(moneropay.ReceivePostResponse{}); err != nil {
				t.Fatal(err)
			}
		}
	})

	r.HandleFunc("/receive/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if err := json.NewEncoder(w).Encode(moneropay.ReceiveGetResponse{}); err != nil {
				t.Fatal(err)
			}
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		}
	})

	r.HandleFunc("/transfer/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(moneropay.TransferGetResponse{})
		case http.MethodPost:
			if err := json.NewDecoder(r.Body).Decode(&moneropay.TransferPostRequest{}); err != nil {
				t.Fatal(err)
			}
			json.NewEncoder(w).Encode(moneropay.TransferPostResponse{})
		}
	})

	server := httptest.NewServer(r)

	mp := MoneropayClient{
		url:    server.URL,
		client: server.Client(),
	}

	_, err := mp.Invoice(big.NewInt(100), "localhost:4420/invoice?id=69")
	if err != nil {
		t.Fatal(err)
	}

	address := testutil.XMRAddress()

	_, err = mp.InvoiceStatus(address)
	if err != nil {
		t.Fatal(err)
	}

	err = mp.DeleteInvoice(address)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mp.Transfer([]provider.Destination{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = mp.TransferStatus("txhash")
	if err != nil {
		t.Fatal(err)
	}
}
