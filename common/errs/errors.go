package errs

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserExists            = errors.New("user already exists")
	ErrComplimentNotFound    = errors.New("compliment not found")
	ErrNoCompliments         = errors.New("compliments not found")
	ErrLimitExceeded         = errors.New("daily limit exceeded")
	ErrImportantDateNotFound = errors.New("important date not found")
	ErrPartnerNotFound       = errors.New("partner not found")
	ErrUserConfigNotFound    = errors.New("user config not found")
)

type ErrBucketEmpty struct {
	Minutes int `json:"minutes"`
}

func (e *ErrBucketEmpty) Error() string {
	return "bucket is empty"
}
