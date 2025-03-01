package validators

import (
	"github.com/go-playground/validator/v10"
	"xrf197ilz35aq2/core/domain"
)

func AuctionTypeValidator(fl validator.FieldLevel) bool {
	auctionType := fl.Field().String()
	return domain.IsValidAuctionType(auctionType)
}
