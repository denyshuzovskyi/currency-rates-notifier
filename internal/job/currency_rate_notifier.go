package job

import (
	"currency-rates-notifier/internal/api/monobank"
	"currency-rates-notifier/internal/config"
	"fmt"
	"github.com/wneessen/go-mail"
	"log/slog"
	"math/rand"
	"text/template"
	"time"
)

type CurrencyRateFetcher interface {
	//todo: return generic currency rate
	FetchUSDToUAHCurrencyRate() (monobank.CurrencyRate, error)
}

type EmailFinder interface {
	GetAllEmails() ([]string, error)
}

type CurrencyRateNotifier struct {
	fetcher     CurrencyRateFetcher
	finder      EmailFinder
	emailClient *mail.Client //todo: create abstraction
	log         *slog.Logger
	cfg         config.Email
}

func NewCurrencyRateNotifier(fetcher CurrencyRateFetcher, finder EmailFinder, emailClient *mail.Client, log *slog.Logger, cfg config.Email) *CurrencyRateNotifier {
	return &CurrencyRateNotifier{fetcher: fetcher, finder: finder, emailClient: emailClient, log: log, cfg: cfg}
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

	textTpl, err := template.New("texttpl").Parse(n.cfg.MessageTemplate)
	if err != nil {
		n.log.Error("failed to parse text template", "error", err)
		return
	}

	var messages []*mail.Msg
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, email := range emails {
		randNum := random.Int31()
		message := mail.NewMsg()
		if err := message.EnvelopeFrom(fmt.Sprintf(n.cfg.EnvelopeFrom, randNum)); err != nil {
			n.log.Error("failed to set ENVELOPE FROM address", "error", err)
			continue
		}
		if err := message.From(n.cfg.From); err != nil {
			n.log.Error("failed to set formatted FROM address", "error", err)
			continue
		}
		if err := message.AddTo(email); err != nil {
			n.log.Error("failed to set formatted TO address", "error", err)
			continue
		}
		message.SetMessageID()
		message.SetDate()
		message.SetBulk()
		message.Subject(n.cfg.Subject)
		if err := message.SetBodyTextTemplate(textTpl, rate); err != nil {
			n.log.Error("failed to add text template to mail body", "error", err)
			continue
		}

		messages = append(messages, message)
	}

	if err := n.emailClient.DialAndSend(messages...); err != nil {
		n.log.Error("failed to deliver mail", "error", err)
		return
	}
	n.log.Info("Bulk mailing successfully delivered.")
}
