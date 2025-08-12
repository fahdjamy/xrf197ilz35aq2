package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"xrf197ilz35aq2/core/domain"
	"xrf197ilz35aq2/internal/exchange"

	"github.com/go-playground/validator/v10"
)

type SessionServ interface {
	CreateSession(request exchange.NewSessionRequest, userFp string, ctx context.Context) (*domain.Session, error)
}

type sessionService struct {
	validate *validator.Validate
	log      slog.Logger
}

func (a *sessionService) CreateSession(request exchange.NewSessionRequest, userFp string, ctx context.Context) (*domain.Session, error) {
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

func NewSessionService(validate *validator.Validate, log slog.Logger) SessionServ {
	return &sessionService{
		log:      log,
		validate: validate,
	}
}
