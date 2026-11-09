package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (h *Handler) NotifyAllImportantDates(w http.ResponseWriter, r *http.Request) {
	var message domain.ImportantDateMessage
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 503 просит API повторить отправку. Если кто-то уже получил сообщение,
	// повтор продублировал бы его, поэтому повторяем только когда не дошло никому
	delivered, temporaryFailure := h.sendNotifications(message)
	if !delivered && temporaryFailure {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (h *Handler) sendNotifications(message domain.ImportantDateMessage) (delivered, temporaryFailure bool) {
	for _, id := range message.UserIDs {
		err := h.ui.Client.SendMessage(id, message.Message)
		if err == nil {
			delivered = true
			continue
		}
		slog.Error("Error sending notification", "chatID", id, "error", err)
		if !errors.Is(err, domain.ErrRecipientUnavailable) {
			temporaryFailure = true
		}
	}
	return delivered, temporaryFailure
}
