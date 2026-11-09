package password

import (
	"errors"
	"testing"

	"github.com/Waycoolers/fmlbot/pkg/errs"
)

func TestValidate(t *testing.T) {
	cases := map[string]error{
		"Passw0rd":                          nil,
		"short1A":                           errs.ErrPasswordTooShort,
		"Aa1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": errs.ErrPasswordTooLong,
		"Пароль1Aaaa":                       errs.ErrPasswordInvalidCharacter,
		"12345678":                          errs.ErrPasswordWithoutLetter,
		"password1":                         errs.ErrPasswordWithoutUpper,
		"PASSWORD1":                         errs.ErrPasswordWithoutLower,
		"Password":                          errs.ErrPasswordWithoutDigit,
	}
	for input, want := range cases {
		if got := Validate(input); !errors.Is(got, want) {
			t.Errorf("Validate(%q) = %v, want %v", input, got, want)
		}
	}
}
