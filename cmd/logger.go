package main

import (
	"fmt"
	"github.com/lmittmann/tint"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"xrf197ilz35aq2/internal"
)

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
	//consoleHandler := slog.NewTextHandler(os.Stdout, &opts)
	logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:   opts.AddSource,
		Level:       opts.Level,
		ReplaceAttr: opts.ReplaceAttr,
		NoColor:     false,
	}))
	logger.Debug(fmt.Sprintf("running in %s mode", env), "logLevel", "DEBUG", "logOutput", "console only")
	return logger, nil
}
