package form

import (
	"github.com/go-playground/validator/v10"
	"github.com/gobugger/gomarket/internal/service/product"
)

func location(fl validator.FieldLevel) bool {
	location := fl.Field().String()
	return product.IsLocationSupported(location)
}
