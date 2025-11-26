package form

import (
	"github.com/go-playground/validator/v10"
	currency_service "github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/product"
	"github.com/gobugger/gomarket/internal/translations"
)

func location(fl validator.FieldLevel) bool {
	location := fl.Field().String()
	return product.IsLocationSupported(location)
}

func xmrAddress(f1 validator.FieldLevel) bool {
	address := f1.Field().String()
	return len(address) == 95
}

func locale(f1 validator.FieldLevel) bool {
	locale := f1.Field().String()
	return translations.ValidLocale(locale)
}

func currency(f1 validator.FieldLevel) bool {
	currency := f1.Field().String()
	return currency_service.Currency(currency).IsSupported()
}
