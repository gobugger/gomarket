package form

import (
	"encoding/gob"
	"github.com/google/uuid"
	"math/big"
)

type RegisterForm struct {
	Username      string `schema:"username" validate:"required,min=3,max=20"`
	Password      string `schema:"password" validate:"required,min=8,max=100"`
	PasswordCheck string `schema:"password_check" validate:"required,eqfield=Password"`
	Captcha
}

type JailForm struct {
	Characters string `schema:"answer" validate:"required,min=1,max=10"`
	Captcha
}

type LoginForm struct {
	Username string `schema:"username" validate:"required,min=3,max=20"`
	Password string `schema:"password" validate:"required,min=8,max=100"`
	Captcha
}

type TwoFactorAuthForm struct {
	Token string `schema:"token" validate:"required,max=100"`
	Captcha
}

type ChangePasswordForm struct {
	Password         string `schema:"password" validate:"required,min=8,max=100"`
	NewPassword      string `schema:"new_password" validate:"required,min=8,max=100"`
	NewPasswordCheck string `schema:"new_password_check" validate:"required,eqfield=NewPassword"`
	Captcha
}

type PGPForm struct {
	PgpKey string `schema:"pgp_key" validate:"required,min=50"`
	Captcha
}

type VendorApplicationForm struct {
	ExistingVendor bool   `schema:"existing_vendor,default:false"`
	Letter         string `schema:"letter" validate:"required,min=20,max=2000"`
	Captcha
}

type VendorLicenseForm struct {
	Captcha
}

type PriceTier struct {
	Quantity int32 `schema:"quantity" validate:"required,gte=0"`
	Price    int64 `schema:"price" validate:"required,gte=0"`
}

type CreateListingForm struct {
	Title       string      `schema:"title" validate:"required,min=3,max=50"`
	Description string      `schema:"description" validate:"required,min=10,max=2000"`
	CategoryID  uuid.UUID   `schema:"category_id" validate:"required,uuid4"`
	PriceTiers  []PriceTier `schema:"price_tiers" validate:"required,min=1,max=10,dive"`
	Inventory   int32       `schema:"inventory" validate:"required,gte=0"`
	ShipsFrom   string      `schema:"ships_from" validate:"required,location"`
	ShipsTo     string      `schema:"ships_to" validate:"required,location"`
	Captcha
}

type UpdateCartForm struct {
	PriceTierID uuid.UUID `schema:"price_tier_id" validate:"required,uuid4"`
	Action      string    `schema:"action" validate:"required,oneof=remove add"`
}

type OrderForm struct {
	DeliveryMethodID uuid.UUID `schema:"delivery_method_id" validate:"required,uuid4"`
	UseWallet        bool      `schema:"use_wallet"`
	Details          string    `schema:"details" validate:"required,min=1,max=2000"`
}

type ProductActionForm struct {
	PriceID uuid.UUID `schema:"price_id" validate:"required,uuid4"`
	Action  string    `schema:"action" validate:"required,oneof=buy_now add_to_cart"`
}

type ProductReview struct {
	ItemID  uuid.UUID `schema:"item_id" validate:"required,uuid4"`
	Grade   int32     `schema:"grade" validate:"required,gte=1,lte=5"`
	Comment string    `schema:"comment" validate:"max=2000"`
}

type ReviewForm struct {
	OrderID        uuid.UUID       `schema:"order_id" validate:"required,uuid4"`
	Grade          int32           `schema:"grade" validate:"required,gte=1,lte=5"`
	Comment        string          `schema:"comment" validate:"max=2000"`
	ProductReviews []ProductReview `schema:"product_reviews" validate:"dive"`
}

type DisputeOfferForm struct {
	OrderID      uuid.UUID `schema:"order_id" validate:"required,uuid4"`
	RefundFactor float64   `schema:"refund_factor" validate:"gte=0,lte=1"`
}

type DisputeOfferResponseForm struct {
	OfferID uuid.UUID `schema:"offer_id" validate:"required,uuid4"`
	Accept  bool      `schema:"accept"`
}

type WithdrawForm struct {
	Address     string     `schema:"address" validate:"required,xmr_address"`
	AmountWhole *big.Float `schema:"amount" validate:"required,gt=0"`
	Captcha
}

type IDForm struct {
	ID uuid.UUID `schema:"id" validate:"required,uuid4"`
}

type ProcessForm struct {
	ID     uuid.UUID `schema:"id" validate:"required,uuid4"`
	Accept bool      `schema:"accept"`
}

type DispatchForm struct {
	OrderID uuid.UUID `schema:"order_id" validate:"required,uuid4"`
}

type TicketForm struct {
	Subject string `schema:"subject" validate:"required,min=3,max=100"`
	Message string `schema:"message" validate:"required,min=5,max=5000"`
	Captcha
}

type TicketResponseForm struct {
	TicketID uuid.UUID `schema:"ticket_id" validate:"required,uuid4"`
	Message  string    `schema:"message" validate:"required,min=3,max=5000"`
	Captcha
}

type SettingsForm struct {
	Locale          string `schema:"locale" validate:"required,locale"`
	Currency        string `schema:"currency" validate:"required,currency"`
	Enable2FA       bool   `schema:"enable_2fa"`
	EnableIncognito bool   `schema:"enable_incognito"`
	Captcha
}

type DeliveryMethod struct {
	Description string  `schema:"description" validate:"required,min=3,max=50"`
	Price       float64 `schema:"price" validate:"required,gte=0"`
}

type DeliveryMethodsForm struct {
	DeliveryMethods []DeliveryMethod `schema:"delivery_methods" validate:"required,max=5,dive"`
	Captcha
}

type UpdateProductForm struct {
	ID        uuid.UUID `schema:"id" validate:"required,uuid4"`
	Inventory int32     `schema:"inventory" validate:"required,gte=0"`
	Delete    bool      `schema:"delete"`
}

type OrderChatForm struct {
	OrderID uuid.UUID `schema:"order_id" validate:"required,uuid4"`
	Message string    `schema:"message" validate:"required,min=1,max=2000"`
	Captcha
}

type UpdateTermsOfService struct {
	TermsOfService string `schema:"tos" validate:"required,min=20,max=5000"`
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
	gob.Register(UpdateTermsOfService{})
}
