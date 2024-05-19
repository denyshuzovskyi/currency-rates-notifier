package monobank

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	log     *slog.Logger
}

func NewClient(baseURL string, log *slog.Logger) *Client {
	return &Client{baseURL: baseURL, log: log}
}

type CurrencyRate struct {
	CurrencyCodeA int32   `json:"currencyCodeA"`
	CurrencyCodeB int32   `json:"currencyCodeB"`
	Date          int64   `json:"date"`
	RateSell      float64 `json:"rateSell"`
	RateBuy       float64 `json:"rateBuy"`
	RateCross     float64 `json:"rateCross"`
	FormattedDate string  `json:"-"`
}

const (
	CurrencyUSD = 840 // ISO 4217 code for USD
	CurrencyUAH = 980 // ISO 4217 code for UAH
)

func findUSDToUAH(rates []CurrencyRate) int {
	for index, rate := range rates {
		if rate.CurrencyCodeA == CurrencyUSD && rate.CurrencyCodeB == CurrencyUAH {
			return index
		}
	}
	return -1
}

func (c *Client) FetchCurrencyRates() ([]CurrencyRate, error) {
	url := fmt.Sprintf("%s/bank/currency", c.baseURL)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.log.Error("failed to close body", "error", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rates []CurrencyRate
	err = json.Unmarshal(body, &rates)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	for i, rate := range rates {
		rates[i].FormattedDate = time.Unix(rate.Date, 0).Format(time.RFC3339)
	}

	return rates, nil
}

func (c *Client) FetchUSDToUAHCurrencyRate() (CurrencyRate, error) {
	rates, err := c.FetchCurrencyRates()
	if err != nil {
		return CurrencyRate{}, err
	}

	USDToUAHIndex := findUSDToUAH(rates)
	if USDToUAHIndex >= 0 {
		return rates[USDToUAHIndex], nil
	}

	return CurrencyRate{}, fmt.Errorf("cannot find usd to uah rate")
}
