package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"xrf197ilz35aq2/internal"
	"xrf197ilz35aq2/validators"
)

func main() {
	env := GetEnvironment()
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

var (
	lock   = &sync.Mutex{}
	logger *slog.Logger
)

func setupLogger(env string) (*slog.Logger, error) {
	lock.Lock()
	defer lock.Unlock()

	var logLevel = new(slog.LevelVar)
	var handlers []slog.Handler
	opts := slog.HandlerOptions{Level: logLevel, AddSource: true}

	if strings.ToUpper(env) == internal.ProductionEnv {
		logLevel.Set(slog.LevelInfo)

		// Console handler for production
		consoleHandler := slog.NewTextHandler(os.Stdout, nil)
		handlers = append(handlers, consoleHandler)

		// File handler for production
		logFile, err := os.OpenFile("xrf-q2.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			slog.Error("failed to open log file", "error", err)
			// Fallback to console-only logging if file cannot be opened
			slog.Warn(fmt.Sprintf("falling back to console-only logging in %s due to file open error", env))
			return nil, err
		}
		fileHandler := slog.NewTextHandler(logFile, &opts)
		handlers = append(handlers, fileHandler)

		wo := io.MultiWriter(os.Stdout, logFile)
		multiWriter := slog.NewTextHandler(wo, &opts)

		logger = slog.New(multiWriter)
		logger.Info("logger setup", "logLevel", "INFO", "logOutputs", "console and file")
		return logger, nil
	}
	// "dev", "test" or any other environment
	logLevel.Set(slog.LevelDebug)
	consoleHandler := slog.NewTextHandler(os.Stdout, &opts)
	logger = slog.New(consoleHandler)
	logger.Debug("running in DEV mode", "logLevel", "DEBUG", "logOutput", "console only")
	return logger, nil
}

func GetEnvironment() string {
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
