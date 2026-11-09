package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/plans"
)

func (c *client) GetPlanOptions(ctx context.Context, chatID int64, request plans.OptionsRequest) (plans.Options, error) {
	var options plans.Options
	err := c.planRequest(ctx, chatID, http.MethodPost, "/plans/options", request, &options)
	return options, err
}

func (c *client) CreatePlan(ctx context.Context, chatID int64, proposal plans.Proposal) (plans.View, error) {
	var view plans.View
	err := c.planRequest(ctx, chatID, http.MethodPost, "/plans", proposal, &view)
	return view, err
}

func (c *client) ListPlans(ctx context.Context, chatID int64) ([]plans.Plan, error) {
	var items []plans.Plan
	err := c.planRequest(ctx, chatID, http.MethodGet, "/plans", nil, &items)
	return items, err
}

func (c *client) GetPlan(ctx context.Context, chatID, planID int64) (plans.View, error) {
	var view plans.View
	err := c.planRequest(ctx, chatID, http.MethodGet, fmt.Sprintf("/plans/%d", planID), nil, &view)
	return view, err
}

func (c *client) ChangePlan(ctx context.Context, chatID, planID int64, change plans.Change) (plans.View, error) {
	var view plans.View
	err := c.planRequest(ctx, chatID, http.MethodPatch, fmt.Sprintf("/plans/%d", planID), change, &view)
	return view, err
}

func (c *client) planRequest(ctx context.Context, chatID int64, method, path string, request, result any) error {
	resp, err := c.doAuthRequest(ctx, method, path, request, chatID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return json.NewDecoder(resp.Body).Decode(result)
	case http.StatusNotFound:
		return errs.ErrPlanNotFound
	case http.StatusTooManyRequests:
		return errs.ErrLimitExceeded
	case http.StatusBadRequest, http.StatusForbidden, http.StatusConflict:
		var body struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(resp.Body).Decode(&body) == nil {
			for _, known := range []error{
				errs.ErrBadRequest, errs.ErrForbidden, errs.ErrPartnerNotFound, errs.ErrPairChanged,
				errs.ErrPlanChanged, errs.ErrPlanTime, errs.ErrPlanPartnerTelegram,
			} {
				if body.Error == known.Error() {
					return known
				}
			}
		}
	}
	return fmt.Errorf("plan request returned status %d", resp.StatusCode)
}
