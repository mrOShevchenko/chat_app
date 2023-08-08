package validators

import (
	"github.com/pkg/errors"
	"regexp"
)

const (
	UsernameMinLength = 4
	UsernameMaxLength = 40
	PasswordMinLength = 8
	PasswordMaxLength = 40
)

var (
	UsernamePattern            = regexp.MustCompile("^[a-zA-Z0-9]+$")
	PasswordLowercasePattern   = regexp.MustCompile("[a-z]+")
	PasswordUppercasePattern   = regexp.MustCompile("[A-Z]+")
	PasswordDigitPattern       = regexp.MustCompile("[0-9]+")
	PasswordSpecialCharPattern = regexp.MustCompile("[!@#$%^&*.?-]+")

	ErrUsernameTooShort          = errors.New("username should be at least 4 characters long")
	ErrUsernameTooLong           = errors.New("username length should be maximum 40 characters long")
	ErrUsernameInvalidCharacters = errors.New("username should contain only lower, upper case latin letters and digits")
	ErrPasswordTooShort          = errors.New("password should be at least 8 characters long")
	ErrPasswordTooLong           = errors.New("password length should be maximum 40 characters long")
	ErrPasswordNoLowercase       = errors.New("password should contain at least one lower case character")
	ErrPasswordNoUppercase       = errors.New("password should contain at least one upper case character")
	ErrPasswordNoDigit           = errors.New("password should contain at least one digit")
	ErrPasswordNoSpecialChar     = errors.New("password should contain at least one special character")
)

// ValidateUsername checks the validity of the provided username
func ValidateUsername(username string) error {
	switch {
	case len(username) < UsernameMinLength:
		return ErrUsernameTooShort
	case len(username) > UsernameMaxLength:
		return ErrUsernameTooLong
	case !UsernamePattern.MatchString(username):
		return ErrUsernameInvalidCharacters
	default:
		return nil
	}
}

// ValidatePassword checks the validity of the provided password
func ValidatePassword(password string) error {
	switch {
	case len(password) < PasswordMinLength:
		return ErrPasswordTooShort
	case len(password) > PasswordMaxLength:
		return ErrPasswordTooLong
	case !PasswordLowercasePattern.MatchString(password):
		return ErrPasswordNoLowercase
	case !PasswordUppercasePattern.MatchString(password):
		return ErrPasswordNoUppercase
	case !PasswordDigitPattern.MatchString(password):
		return ErrPasswordNoDigit
	case !PasswordSpecialCharPattern.MatchString(password):
		return ErrPasswordNoSpecialChar
	default:
		return nil
	}
}
