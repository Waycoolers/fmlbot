package usecases

import "github.com/Waycoolers/fmlbot/services/api/internal/domain"

type UseCase struct {
	users          domain.UsersRepo
	userConfig     domain.UserConfigRepo
	compliments    domain.ComplimentsRepo
	importantDates domain.ImportantDatesRepo
	scheduler      domain.SchedulerRepo
}

func New(repos *domain.Repos) *UseCase {
	return &UseCase{
		users:          repos.Users,
		userConfig:     repos.UserConfig,
		compliments:    repos.Compliments,
		importantDates: repos.ImportantDates,
		scheduler:      repos.Scheduler,
	}
}
