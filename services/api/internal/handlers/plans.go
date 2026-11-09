package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/middlewares"
	"github.com/Waycoolers/fmlbot/pkg/plans"
)

func writePlanError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, errs.ErrBadRequest), errors.Is(err, errs.ErrPlanTime):
		status = http.StatusBadRequest
	case errors.Is(err, errs.ErrPlanNotFound):
		status = http.StatusNotFound
	case errors.Is(err, errs.ErrPairChanged), errors.Is(err, errs.ErrPlanChanged), errors.Is(err, errs.ErrPlanPartnerTelegram):
		status = http.StatusConflict
	case errors.Is(err, errs.ErrForbidden), errors.Is(err, errs.ErrPartnerNotFound):
		status = http.StatusForbidden
	default:
		slog.Error("Plan request failed", "error", err)
		err = errors.New("internal error")
	}
	sendJson(w, status, map[string]string{"error": err.Error()})
}

func planUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok || userID <= 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}

func (h *Handler) GetPlanOptions(w http.ResponseWriter, r *http.Request) {
	userID, ok := planUser(w, r)
	if !ok {
		return
	}
	var request plans.OptionsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || !request.Pair.Valid() || !request.Prompt.Valid() {
		writePlanError(w, errs.ErrBadRequest)
		return
	}
	if !h.ideaLimiter.Allow(strconv.FormatInt(userID, 10)) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}
	options, err := h.uc.GetPlanOptions(r.Context(), userID, request)
	if errors.Is(err, errs.ErrPartnerNotFound) || errors.Is(err, errs.ErrPairChanged) {
		writePlanError(w, err)
		return
	}
	if err != nil {
		slog.Warn("Plan generation failed", "error", err)
		http.Error(w, "could not generate plans", http.StatusBadGateway)
		return
	}
	sendJson(w, http.StatusOK, options)
}

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := planUser(w, r)
	if !ok {
		return
	}
	var request plans.Proposal
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writePlanError(w, errs.ErrBadRequest)
		return
	}
	if !h.pairLimiter.Allow("plan:" + strconv.FormatInt(userID, 10)) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}
	plan, err := h.uc.CreatePlan(r.Context(), userID, request)
	if err != nil {
		writePlanError(w, err)
		return
	}
	sendJson(w, http.StatusCreated, plans.View{Plan: plan, ViewerID: userID})
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	userID, ok := planUser(w, r)
	if !ok {
		return
	}
	items, err := h.uc.ListPlans(r.Context(), userID)
	if err != nil {
		writePlanError(w, err)
		return
	}
	sendJson(w, http.StatusOK, items)
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := planUser(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writePlanError(w, errs.ErrPlanNotFound)
		return
	}
	plan, err := h.uc.GetPlan(r.Context(), userID, id)
	if err != nil {
		writePlanError(w, err)
		return
	}
	sendJson(w, http.StatusOK, plans.View{Plan: plan, ViewerID: userID})
}

func (h *Handler) ChangePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := planUser(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writePlanError(w, errs.ErrPlanNotFound)
		return
	}
	var request plans.Change
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writePlanError(w, errs.ErrBadRequest)
		return
	}
	if request.Action == plans.Reschedule && !h.pairLimiter.Allow("plan:"+strconv.FormatInt(userID, 10)) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}
	plan, err := h.uc.ChangePlan(r.Context(), userID, id, request)
	if err != nil {
		writePlanError(w, err)
		return
	}
	sendJson(w, http.StatusOK, plans.View{Plan: plan, ViewerID: userID})
}
