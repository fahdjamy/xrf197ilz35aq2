package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"xrf197ilz35aq2/core/domain"
	"xrf197ilz35aq2/internal/exchange"
)

type SessionServ interface {
	NewAuction(request exchange.NewSessionRequest, userFp string, ctx context.Context) (*domain.Session, error)
}

type sessionService struct {
	validate *validator.Validate
}

func (a *sessionService) NewAuction(request exchange.NewSessionRequest, userFp string, ctx context.Context) (*domain.Session, error) {
	err := a.validateRequest(request)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (a *sessionService) validateRequest(req exchange.NewSessionRequest) error {
	err := a.validate.Struct(req)
	if err != nil {
		var validationErrors validator.ValidationErrors
		errors.As(err, &validationErrors)
		invalidFields := make([]string, 0)
		for _, validationError := range validationErrors {
			invalidFields = append(invalidFields, validationError.Field())
		}
		return errors.New(fmt.Sprintf("invalid fields [%s]", invalidFields))
	}
	return nil
}

func NewAuctionService() SessionServ {
	return &sessionService{}
}
