package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"xrf197ilz35aq2/internal"
	"xrf197ilz35aq2/server/grpc"
	"xrf197ilz35aq2/server/socket"
	"xrf197ilz35aq2/storage"
	"xrf197ilz35aq2/storage/postgres"
	"xrf197ilz35aq2/storage/redis"
	"xrf197ilz35aq2/storage/timescale"
	"xrf197ilz35aq2/validators"
)

const gRPCPortAddress = ":50052"

func main() {
	env := getAppEnv()
	config, err := internal.GetConfig(strings.ToLower(env))

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
	_, err = setTimescaleDB(config, *logger)
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

	// /////// Set up redis client
	redisClient, err := storage.NewRedisClient(context.Background(), config.Redis)
	if err != nil {
		logger.Error("failed to create redis client", "err", err)
		return
	}

	// /////// Set up postgres client
	pgPool, err := storage.NewPGConnection(context.Background(), config.Postgres.DatabaseURL, *logger)
	if err != nil {
		logger.Error("failed to create postgres client", "err", err)
		return
	}
	defer pgPool.Close()

	cacheClient := redis.CacheClients{
		BidClient: redis.NewBidCache(*logger, redisClient),
	}
	allRepos := postgres.Repositories{
		BidRepository:     postgres.NewBidRepo(pgPool.Pool, *logger),
		SessionRepository: postgres.NewSessionRepository(pgPool.Pool, *logger),
	}

	runApp(err, logger, cacheClient, allRepos)
}

func runApp(err error, logger *slog.Logger, cacheClient redis.CacheClients, allRepos postgres.Repositories) {
	/////// 1. Create a TCP listener on the specified port
	listener, err := net.Listen("tcp", gRPCPortAddress)
	if err != nil {
		logger.Error("failed to start listening on gRPC port", "port", gRPCPortAddress, "err", err)
		return
	}

	// Context that is to be canceled when a shutdown signal is received.
	cancellableCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create an err-group tied to our cancellable context.
	g, gCtx := errgroup.WithContext(cancellableCtx)

	/////// 2. Create a websocket hub and start it ---
	hub := socket.NewHub(*logger)
	g.Go(func() error {
		return hub.Run(gCtx)
	})

	/////// 3. start websocket server in a separate go routine
	// TODO: IN production, use ListenAndServeTLS
	server := &http.Server{
		Addr: ":8082",
	}

	g.Go(func() error {
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			// TODO: Authenticate the user here before upgrading the connection.
			socket.ServeWS(hub, w, r, *logger)
		})

		logger.Info("starting websocket http server on port 8082")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("starting http server error", err)
			return fmt.Errorf("starting http server: %w", err)
		}
		return nil
	})

	//////// 4. start the gRPC server in a go routine
	grpcServer, err := grpc.NewGRPCSrv(*logger, cacheClient, allRepos, hub)
	g.Go(func() error {
		logger.Info("starting gRPC server", "port", gRPCPortAddress)
		if err = grpcServer.Serve(listener); err != nil {
			logger.Error("Failed to serve gRPC server", "err", err)
			return fmt.Errorf("starting gRPC server: %w", err)
		}
		return nil
	})

	healthBeatCheck(*logger)

	// gracefully shut down the application if one of the servers fails, i.e., when the context is canceled.
	g.Go(func() error {
		<-cancellableCtx.Done() // Block until the context is canceled.
		logger.Info("!!!!!! shutting down xrf197ilz35aq !!!!!!")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown xrf197ilz35aq", "err", err)
		}
		return nil
	})

	logger.Info("**** started xrf197ilz35aq (ii) app successfully *****")

	if err := g.Wait(); err != nil {
		// If Wait() returns an error, it means one of the goroutines (i.e., servers) failed to start.
		logger.Error("failed to start xrf197ilz35aq", "err", err)
	} else {
		logger.Info("**** xrf197ilz35aq (ii) existing graceful *****")
	}
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

func setTimescaleDB(config *internal.Config, logger slog.Logger) (*storage.TimescaleDB, error) {
	ctx := context.Background()
	timescaleDBUrl := config.TimescaleDB.DatabaseURL
	if timescaleDBUrl == "" {
		return nil, fmt.Errorf("failed to load timescaleDBUrl")
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
	err = timescale.MigrateTimescaleTables(migrateCtx, timescalePool.Pool, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate timescaleDB: tables :: err=%w", err)
	}
	return timescalePool, nil
}

func healthBeatCheck(logger slog.Logger) {
	for range time.Tick(time.Second * 60) {
		logger.Info("event=appHealthCheck", "message", "healthy")
	}
}
