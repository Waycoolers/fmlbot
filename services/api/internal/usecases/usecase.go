package usecases

import "github.com/Waycoolers/fmlbot/services/api/internal/domain"

type UseCase struct {
	repo *domain.Repo
}

func New(repo *domain.Repo) *UseCase {
	return &UseCase{repo: repo}
}
