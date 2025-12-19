package ui

import (
	"encoding/gob"
	"github.com/gobugger/gomarket/internal/form"
	"github.com/google/uuid"
	"math/big"
)

type Form struct {
	Values      map[string]string
	FieldErrors form.FieldErrors
}

type Note struct {
	Index   int
	Message string
	IsError bool
}

type SiteConfig struct {
	Name  string
	Onion string
}

type SiteStats struct {
	NumVendors int64
	NumUsers   int64
}

type UserSettings struct {
	IncognitoEnabled bool
	TwofaEnabled     bool
	Lang             string
	Currency         string
}

type AuthLevel int

const (
	AuthLevelDefault AuthLevel = iota
	AuthLevelCustomer
	AuthLevelVendor
	AuthLevelAdmin
)

type TemplateContext struct {
	UID                    uuid.UUID
	Username               string
	PgpKey                 string
	AuthLevel              AuthLevel
	BalancePico            *big.Int
	Settings               UserSettings
	Notes                  []Note
	NumUnseenNotifications int
	NumCartItems           int
	Config                 SiteConfig
	Stats                  SiteStats
	CsrfField              string
	CaptchaSrc             string
	Form                   Form
}

func (tc *TemplateContext) IsAuthenticated() bool {
	return tc != nil && tc.AuthLevel >= AuthLevelCustomer
}

func (tc *TemplateContext) IsCustomer() bool {
	return tc != nil && tc.AuthLevel == AuthLevelCustomer
}

func (tc *TemplateContext) IsVendor() bool {
	return tc != nil && tc.AuthLevel == AuthLevelVendor
}

func init() {
	gob.Register(Form{})
	gob.Register(Note{})
	gob.Register([]Note{})
}
