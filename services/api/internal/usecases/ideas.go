package usecases

import (
	"context"

	"github.com/Waycoolers/fmlbot/pkg/errs"
)

func (uc *UseCase) GetLeisureIdea(ctx context.Context, userID int64, location, level, budget, context string) (string, error) {
	partnerID, err := uc.users.GetPartnerID(ctx, userID)
	if err != nil {
		return "", err
	}
	if partnerID == 0 {
		return "", errs.ErrPartnerNotFound
	}
	return uc.ai.GetLeisureIdea(ctx, location, level, budget, context)
}
