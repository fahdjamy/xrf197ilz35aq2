package main

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"xrf197ilz35aq2/internal"
	"xrf197ilz35aq2/storage"
	"xrf197ilz35aq2/storage/timescale"
	"xrf197ilz35aq2/validators"
)

func main() {
	env := getAppEnv()
	config, err := internal.NewConfig(strings.ToLower(env))
	_, err = internal.GetConfig(strings.ToLower(env))

	if err != nil {
		fmt.Println("failed to load config:", err)
		return
	}

	logger, err := internal.SetupLogger(env, config.Log)
	if err != nil {
		fmt.Printf("Failed to setup logger: %v\n", err)
		return
	}

	// setup Databases
	_, err = setTimescaleDB(config, logger)
	if err != nil {
		logger.Error("Failed to setup timescaleDB: %s\n", "err", err)
		return
	}

	var validate *validator.Validate
	validate = validator.New()

	err = validate.RegisterValidation("auctionType", validators.AuctionTypeValidator)
	if err != nil {
		logger.Error("Register auctionType validation error: %s\n", "err", err)
		return
	}
	logger.Info("started xrf197ilz35aq")

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal // Block until the shutdown signal is received
	logger.Info("\n shuttingDown xrf197ilz35aq")
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

func setTimescaleDB(config *internal.Config, logger *slog.Logger) (*storage.TimescaleDB, error) {
	ctx := context.Background()
	timescaleDBUrl, err := config.TimescaleDB.GetDdURL()
	if err != nil {
		return nil, fmt.Errorf("failed to load timescaleDBUrl :: err=%w", err)
	}

	connTSDBCtx, cancelFunc := context.WithTimeout(ctx, 10*time.Second)
	defer cancelFunc()
	timescalePool, err := storage.GetTimescaleDBConn(connTSDBCtx, timescaleDBUrl, logger, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to timescaleDB :: err=%w", err)
	}
	err = timescalePool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("timescaleDB connection ping failed :: err=%w", err)
	}

	migrateCtx, cancelFunc := context.WithTimeout(ctx, 20000*time.Second)
	defer cancelFunc()
	err = timescale.MigrateTimescaleTables(migrateCtx, timescalePool.Pool)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate timescaleDB: tables :: err=%w", err)
	}
	return timescalePool, nil
}
