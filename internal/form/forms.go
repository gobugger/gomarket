package form

import (
	"encoding/gob"
	"github.com/google/uuid"
)

type RegisterForm struct {
	Username      string `schema:"username" validate:"required,min=3,max=50"`
	Password      string `schema:"password" validate:"required,min=8,max=100"`
	PasswordCheck string `schema:"password_check" validate:"required,eqfield=Password"`
	Captcha
}

type JailForm struct {
	Characters string `schema:"answer"`
	Captcha
}

type LoginForm struct {
	Username string `schema:"username" validate:"required,min=3,max=50"`
	Password string `schema:"password" validate:"required,min=8,max=100"`
	Captcha
}

type TwoFactorAuthForm struct {
	Token string `schema:"token"`
	Captcha
}

type ChangePasswordForm struct {
	Password         string `schema:"password"`
	NewPassword      string `schema:"new_password"`
	NewPasswordCheck string `schema:"new_password_check"`
	Captcha
}

type PGPForm struct {
	PgpKey string `schema:"pgp_key"`
	Captcha
}

type VendorApplicationForm struct {
	ExistingVendor bool   `schema:"existing_vendor,default:false"`
	Letter         string `schema:"letter"`
	Captcha
}

type VendorLicenseForm struct {
	Captcha
}

type PriceTier struct {
	Quantity int32 `schema:"quantity" validate:"gte=0"`
	Price    int64 `schema:"price" validate:"gte=0"`
}

type CreateListingForm struct {
	Title       string      `schema:"title" validate:"required,max=50"`
	Description string      `schema:"description" validate:"required,max=1000"`
	CategoryID  uuid.UUID   `schema:"category_id" validate:"required"`
	PriceTiers  []PriceTier `schema:"price_tiers" validate:"required,min=1,max=10,dive"`
	Inventory   int32       `schema:"inventory" validate:"gte=0"`
	ShipsFrom   string      `schema:"ships_from" validate:"required,location"`
	ShipsTo     string      `schema:"ships_to" validate:"required,location"`
	Captcha
}

type UpdateCartForm struct {
	PriceTierID uuid.UUID `schema:"price_tier_id" validate:"required"`
	Action      string    `schema:"action" validate:"required,oneof=remove add"`
}

type OrderForm struct {
	DeliveryMethodID uuid.UUID `schema:"delivery_method_id" validate:"required"`
	UseWallet        bool      `schema:"use_wallet"`
	Details          string    `schema:"details" validate:"required"`
}

type ProductActionForm struct {
	PriceID uuid.UUID `schema:"price_id" validate:"required"`
	Action  string    `schema:"action" validate:"required,oneof=buy_now add_to_cart"`
}

type ProductReview struct {
	ItemID  uuid.UUID `schema:"item_id"`
	Grade   int32     `schema:"grade"`
	Comment string    `schema:"comment"`
}

type ReviewForm struct {
	OrderID        uuid.UUID       `schema:"order_id"`
	Grade          int32           `schema:"grade"`
	Comment        string          `schema:"comment"`
	ProductReviews []ProductReview `schema:"product_reviews" validate:"dive"`
}

type DisputeOfferForm struct {
	OrderID      uuid.UUID `schema:"order_id"`
	RefundFactor float64   `schema:"refund_factor"`
}

type DisputeOfferResponseForm struct {
	OfferID uuid.UUID `schema:"offer_id"`
	Accept  bool      `schema:"accept"`
}

type WithdrawForm struct {
	Address   string  `schema:"address"`
	AmountXMR float64 `schema:"amount_xmr"`
	Captcha
}

type IDForm struct {
	ID uuid.UUID `schema:"id"`
}

type ProcessForm struct {
	ID     uuid.UUID `schema:"id"`
	Accept bool      `schema:"accept"`
}

type DispatchForm struct {
	OrderID uuid.UUID `schema:"order_id"`
}

type TicketForm struct {
	Subject string `schema:"subject"`
	Message string `schema:"message"`
	Captcha
}

type TicketResponseForm struct {
	TicketID uuid.UUID `schema:"ticket_id"`
	Message  string    `schema:"message"`
	Captcha
}

type SettingsForm struct {
	Locale          string `schema:"locale"`
	Currency        string `schema:"currency"`
	Enable2FA       bool   `schema:"enable_2fa"`
	EnableIncognito bool   `schema:"enable_incognito"`
	Captcha
}

type DeliveryMethod struct {
	Description string  `schema:"description" validate:"max=50"`
	Price       float64 `schema:"price" validate:"gte=0"`
}

type DeliveryMethodsForm struct {
	DeliveryMethods []DeliveryMethod `schema:"delivery_methods" validate:"max=5,dive"`
	Captcha
}

type UpdateProductForm struct {
	ID        uuid.UUID `schema:"id"`
	Inventory int32     `schema:"inventory"`
	Delete    bool      `schema:"delete,default:false"`
}

type OrderChatForm struct {
	OrderID uuid.UUID `schema:"order_id"`
	Message string    `schema:"message"`
	Captcha
}

type UpdateVendorProfileForm struct {
	VendorInfo string `schema:"info"`
	Captcha
}

func init() {
	gob.Register(uuid.UUID{})
	gob.Register(JailForm{})
	gob.Register(RegisterForm{})
	gob.Register(LoginForm{})
	gob.Register(ChangePasswordForm{})
	gob.Register(PGPForm{})
	gob.Register(VendorApplicationForm{})
	gob.Register(CreateListingForm{})
	gob.Register(OrderForm{})
	gob.Register(ReviewForm{})
	gob.Register(DisputeOfferForm{})
	gob.Register(DisputeOfferResponseForm{})
	gob.Register(WithdrawForm{})
	gob.Register(ProcessForm{})
	gob.Register(DispatchForm{})
	gob.Register(TicketForm{})
	gob.Register(TicketResponseForm{})
	gob.Register(IDForm{})
	gob.Register(SettingsForm{})
	gob.Register(UpdateProductForm{})
	gob.Register(OrderChatForm{})
	gob.Register(UpdateVendorProfileForm{})
}
