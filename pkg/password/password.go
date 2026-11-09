package password

import "github.com/Waycoolers/fmlbot/pkg/errs"

// Validate проверяет пароль по единой для всех сервисов политике.
func Validate(password string) error {
	if len([]rune(password)) < 8 {
		return errs.ErrPasswordTooShort
	}
	if len([]rune(password)) > 32 {
		return errs.ErrPasswordTooLong
	}

	var hasLetter, hasUpper, hasLower, hasNumber bool

	for _, char := range password {
		if char < '!' || char > '~' {
			return errs.ErrPasswordInvalidCharacter
		}

		switch {
		case char >= 'A' && char <= 'Z':
			hasLetter = true
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLetter = true
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}

	if !hasLetter {
		return errs.ErrPasswordWithoutLetter
	}
	if !hasUpper {
		return errs.ErrPasswordWithoutUpper
	}
	if !hasLower {
		return errs.ErrPasswordWithoutLower
	}
	if !hasNumber {
		return errs.ErrPasswordWithoutDigit
	}
	return nil
}
