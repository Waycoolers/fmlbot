package usecases

import "context"

func (uc *UseCase) GetLeisureIdea(ctx context.Context, location, level, budget, context string) (string, error) {
	return uc.ai.GetLeisureIdea(ctx, location, level, budget, context)
}
