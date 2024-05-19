package handler

import (
	"currency-rates-notifier/internal/api/monobank"
	"currency-rates-notifier/internal/lib/httputil"
	"log/slog"
	"net/http"
)

type CurrencyRateHandler struct {
	monobankClient *monobank.Client
	log            *slog.Logger
}

func NewCurrencyRateHandler(monobankClient *monobank.Client, log *slog.Logger) *CurrencyRateHandler {
	return &CurrencyRateHandler{monobankClient: monobankClient, log: log}
}

func (h *CurrencyRateHandler) GetCurrencyRate(w http.ResponseWriter, r *http.Request) {

	rate, err := h.monobankClient.FetchUSDToUAHCurrencyRate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := httputil.WriteJSON(w, rate); err != nil {
		h.log.Error("failed to write a currencyRate", "error", err)
		return
	}
}
