package job

import (
	"currency-rates-notifier/internal/api/monobank"
	"log/slog"
)

type CurrencyRateFetcher interface {
	FetchUSDToUAHCurrencyRate() (monobank.CurrencyRate, error)
}

type EmailFinder interface {
	GetAllEmails() ([]string, error)
}

type CurrencyRateNotifier struct {
	fetcher CurrencyRateFetcher
	finder  EmailFinder
	log     *slog.Logger
}

func NewCurrencyRateNotifier(fetcher CurrencyRateFetcher, finder EmailFinder, log *slog.Logger) *CurrencyRateNotifier {
	return &CurrencyRateNotifier{fetcher: fetcher, finder: finder, log: log}
}

func (n *CurrencyRateNotifier) SendEmailToSubscribers() {
	rate, err := n.fetcher.FetchUSDToUAHCurrencyRate()
	if err != nil {
		n.log.Error("failed fetch currency rate", "error", err)
	}

	emails, err := n.finder.GetAllEmails()
	if err != nil {
		n.log.Error("failed to get emails", "error", err)
	}

	for _, email := range emails {
		// send email
		_, _ = rate, email
	}
}
