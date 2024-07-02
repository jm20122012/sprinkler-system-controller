package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sprinkler-controller-service/internal/api"
	"sprinkler-controller-service/internal/config"
	cs "sprinkler-controller-service/internal/controllerservice"
	"sprinkler-controller-service/internal/utils"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	slog.Info("Loading .env...")
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
		panic(err)
	}

	// Load app config
	slog.Info("Loading app config...")
	cfg, err := config.LoadConfig("./config.json")
	if err != nil {
		slog.Error("Erorr loading config.json", "error", err)
		panic(err)
	}

	// Setup logger
	var logger *slog.Logger
	slog.Info("Setting debug level", "level", cfg.AppConfig.DebugLevel)
	switch cfg.AppConfig.DebugLevel {
	case "DEBUG":
		logger = utils.CreateLogger(slog.LevelDebug)
	case "INFO":
		logger = utils.CreateLogger(slog.LevelInfo)
	case "WARNING":
		logger = utils.CreateLogger(slog.LevelWarn)
	case "ERROR":
		logger = utils.CreateLogger(slog.LevelError)
	default:
		logger = utils.CreateLogger(slog.LevelInfo)
	}

	logger.Debug("Config loaded", "config", cfg)

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func(cancel context.CancelFunc) {
		<-c
		logger.Info("Ctrl+C pressed, cancelling context...")
		cancel()
	}(cancel)

	var apiHndlr api.IApiHandler
	if cfg.AppConfig.DryRun {
		apiHndlr = api.NewDryRunApiHandler(ctx, logger, cfg.AppConfig.ApiUrl)
	} else {
		apiHndlr = api.NewApiHandler(ctx, logger, cfg.AppConfig.ApiUrl)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	controllerService := cs.NewControllerService(
		ctx,
		wg,
		logger,
		cfg,
		apiHndlr,
	)

	go controllerService.Run()

	wg.Wait()
	logger.Info("Exiting...")
}
