package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"os"
	"os/signal"
	"syscall"
	"xrf197ilz35aq2/internal"
	"xrf197ilz35aq2/validators"
)

func main() {
	env := getAppEnv()
	logger, err := setupLogger(env)
	if err != nil {
		fmt.Printf("Failed to setup logger: %v\n", err)
		return
	}

	var validate *validator.Validate
	validate = validator.New()

	err = validate.RegisterValidation("auctionType", validators.AuctionTypeValidator)
	if err != nil {
		logger.Error("Register auctionType validation error: %s\n", err)
		return
	}
	logger.Info("starting xrf197ilz35aq")

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal // Block until shutdown signal is received
	logger.Info("shuttingDown xrf197ilz35aq")
}

func getAppEnv() string {
	env, ok := os.LookupEnv(internal.Environment)
	if !ok || env == "" {
		env = internal.DevelopEnv
	}

	switch env {
	case internal.StagingEnv:
		return "STAGING"
	case internal.ProductionEnv, internal.LiveEnv:
		return internal.LiveEnv
	default:
		return internal.DevelopEnv
	}
}
