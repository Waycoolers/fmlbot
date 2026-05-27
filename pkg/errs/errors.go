package errs

import "errors"

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrUserExists               = errors.New("user already exists")
	ErrComplimentNotFound       = errors.New("compliment not found")
	ErrNoCompliments            = errors.New("compliments not found")
	ErrLimitExceeded            = errors.New("daily limit exceeded")
	ErrImportantDateNotFound    = errors.New("important date not found")
	ErrPartnerNotFound          = errors.New("partner not found")
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
