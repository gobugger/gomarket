package model

type Health struct {
	Status  int `json:"status"`
	Service struct {
		Database bool `json:"database"`
	}
}
