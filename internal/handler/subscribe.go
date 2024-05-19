package handler

import (
	"currency-rates-notifier/internal/storage"
	"errors"
	"log/slog"
	"net/http"
)

type EmailSaver interface {
	SaveEmail(email string) error
}

type SubscriptionHandler struct {
	saver EmailSaver
	log   *slog.Logger
}

func NewSubscriptionHandler(saver EmailSaver, log *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{saver: saver, log: log}
}

func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")

	err = h.saver.SaveEmail(email)
	if errors.Is(err, storage.EmailExists) {
		w.WriteHeader(http.StatusConflict)

		return
	}
}
