package usecases

import (
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

type UseCase struct {
	users          domain.UsersRepo
	userConfig     domain.UserConfigRepo
	compliments    domain.ComplimentsRepo
	importantDates domain.ImportantDatesRepo
	scheduler      domain.SchedulerRepo
	fcm            domain.FCMRepo
	plans          domain.PlansRepo
	ai             domain.AIClient
	auth           domain.AuthClient
}

func New(repos *domain.Repos, ai domain.AIClient, auth domain.AuthClient) *UseCase {
	return &UseCase{
		users:          repos.Users,
		userConfig:     repos.UserConfig,
		compliments:    repos.Compliments,
		importantDates: repos.ImportantDates,
		scheduler:      repos.Scheduler,
		fcm:            repos.FCM,
		plans:          repos.Plans,
		ai:             ai,
		auth:           auth,
	}
}
