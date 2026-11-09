package errs

import "errors"

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrTelegramLinkInvalid      = errors.New("invalid or expired Telegram link")
	ErrTelegramAlreadyLinked    = errors.New("Telegram account already linked")
	ErrUserExists               = errors.New("user already exists")
	ErrComplimentDelivery       = errors.New("compliment delivery failed")
	ErrComplimentNotFound       = errors.New("compliment not found")
	ErrNoCompliments            = errors.New("compliments not found")
	ErrLimitExceeded            = errors.New("daily limit exceeded")
	ErrImportantDateNotFound    = errors.New("important date not found")
	ErrRecipientUnavailable     = errors.New("recipient unavailable")
	ErrPairChanged              = errors.New("pair changed")
	ErrPartnerNotFound          = errors.New("partner not found")
	ErrPlanNotFound             = errors.New("plan not found")
	ErrPlanChanged              = errors.New("plan changed")
	ErrPlanTime                 = errors.New("invalid plan time")
	ErrPlanPartnerTelegram      = errors.New("partner has no linked Telegram account")
	ErrUserConfigNotFound       = errors.New("user config not found")
	ErrWrongPassword            = errors.New("wrong password")
	ErrBadRequest               = errors.New("bad request")
	ErrPasswordTooShort         = errors.New("password too short")
	ErrPasswordTooLong          = errors.New("password too long")
	ErrPasswordInvalidCharacter = errors.New("password invalid")
	ErrPasswordWithoutLetter    = errors.New("password without letter")
	ErrPasswordWithoutUpper     = errors.New("password without uppercase")
	ErrPasswordWithoutLower     = errors.New("password without lowercase")
	ErrPasswordWithoutDigit     = errors.New("password without digit")
	ErrUsernameIsAlreadyTaken   = errors.New("username is already taken")
	ErrTokenNotFound            = errors.New("token not found")
	ErrTokenInvalid             = errors.New("invalid token")
	ErrForbidden                = errors.New("forbidden")
	ErrInvalidImportantDate     = errors.New("invalid important date")
	ErrPartnerRequestNotFound   = errors.New("partner request not found")
	ErrCannotPartnerYourself    = errors.New("cannot partner yourself")
	ErrAlreadyHasPartner        = errors.New("already has partner")
	ErrPartnerAlreadyHasPartner = errors.New("partner already has partner")
)

type ErrBucketEmpty struct {
	Minutes int `json:"minutes"`
}

func (e *ErrBucketEmpty) Error() string {
	return "bucket is empty"
}
