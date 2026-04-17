package handlers

import (
	"encoding/json"
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

	h.sendNotifications(message)
}

func (h *Handler) sendNotifications(message domain.ImportantDateMessage) {
	for _, id := range message.UserIDs {
		h.Reply(id, message.Message)
	}
}
