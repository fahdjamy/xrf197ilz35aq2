package main

import (
	"fmt"
	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"xrf197ilz35aq2/internal"
)

var (
	lock    = &sync.Mutex{}
	sLogger *slog.Logger
)

func setupLogger(env string, config internal.LogConfig) (*slog.Logger, error) {
	lock.Lock()
	defer lock.Unlock()

	var logLevel = new(slog.LevelVar)
	opts := slog.HandlerOptions{Level: logLevel, AddSource: true}

	if strings.ToUpper(env) == internal.ProductionEnv {
		fileOutput := config.OutputFile
		logFileWriter := &lumberjack.Logger{
			Filename:   fileOutput,
			MaxSize:    100, // megabytes
			MaxBackups: 5,
			MaxAge:     10, // days
			LocalTime:  true,
			Compress:   true,
		}

		logLevel.Set(slog.LevelInfo)

		wo := io.MultiWriter(os.Stdout, logFileWriter)
		multiWriter := slog.NewTextHandler(wo, &opts)

		sLogger = slog.New(multiWriter)
		sLogger.Info("logger setup", "logLevel", "INFO", "logOutputs", "console and file")
		return sLogger, nil
	}

	logLevel.Set(slog.LevelDebug) // "dev", "test" or any other environment
	sLogger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:       opts.Level,
		AddSource:   opts.AddSource,
		ReplaceAttr: opts.ReplaceAttr,
	}))
	sLogger.Debug(fmt.Sprintf("running in %s mode", env), "logLevel", "DEBUG", "logOutput", "console only")
	return sLogger, nil
}
