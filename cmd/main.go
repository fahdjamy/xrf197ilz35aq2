package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"xrf197ilz35aq2/validators"
)

func main() {
	var validate *validator.Validate
	validate = validator.New()

	err := validate.RegisterValidation("auctionType", validators.AuctionTypeValidator)
	if err != nil {
		fmt.Printf("Register auctionType validation error: %s\n", err)
		return
	}
	fmt.Println(".....xrf197ilz35aq || started.....")
}
