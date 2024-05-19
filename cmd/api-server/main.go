package main

import (
	"currency-rates-notifier/internal/api/monobank"
	"currency-rates-notifier/internal/config"
	"currency-rates-notifier/internal/handler"
	"currency-rates-notifier/internal/storage/sqlite"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg := config.ReadConfig("./config/local.yaml")
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	monobankClient := monobank.NewClient(cfg.Monobank.API.URL, log)

	storage, err := sqlite.New("./storage.db")
	if err != nil {
		log.Error("failed to init storage", "error", err)
		os.Exit(1)
	}

	router := http.NewServeMux()
	currencyRateHandler := handler.NewCurrencyRateHandler(monobankClient, log)
	subscriptionHandler := handler.NewSubscriptionHandler(storage, log)

	router.HandleFunc("GET /rate", currencyRateHandler.GetCurrencyRate)
	router.HandleFunc("POST /subscribe", subscriptionHandler.Subscribe)

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler: router,
	}

	log.Info("starting server", "host", cfg.HTTPServer.Host, "port", cfg.HTTPServer.Port)

	err = server.ListenAndServe()
	if err != nil {
		log.Error("failed to start server", "error", err)
		return
	}
}
