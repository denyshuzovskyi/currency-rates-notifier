package handler

import (
	"currency-rates-notifier/internal/api/monobank"
	"currency-rates-notifier/internal/lib/httputil"
	"log/slog"
	"net/http"
)

type CurrencyRateFetcher interface {
	FetchUSDToUAHCurrencyRate() (monobank.CurrencyRate, error)
}

type CurrencyRateHandler struct {
	fetcher CurrencyRateFetcher
	log     *slog.Logger
}

func NewCurrencyRateHandler(fetcher CurrencyRateFetcher, log *slog.Logger) *CurrencyRateHandler {
	return &CurrencyRateHandler{fetcher: fetcher, log: log}
}

func (h *CurrencyRateHandler) GetCurrencyRate(w http.ResponseWriter, r *http.Request) {

	rate, err := h.fetcher.FetchUSDToUAHCurrencyRate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := httputil.WriteJSON(w, rate); err != nil {
		h.log.Error("failed to write a currencyRate", "error", err)
		return
	}
}
