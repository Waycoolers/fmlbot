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

	if len(message.UserIDs) != 1 {
		http.Error(w, "expected one recipient", http.StatusBadRequest)
		return
	}
	h.sendUpdate(w, message.UserIDs[0], message.Message)
}
