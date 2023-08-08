package validators

func ValidateQuery(query string) error {
	switch {
	case len(query) < UsernameMinLength:
		return ErrUsernameTooShort
	case len(query) > UsernameMaxLength:
		return ErrUsernameTooLong
	case !UsernamePattern.MatchString(query):
		return ErrUsernameInvalidCharacters
	default:
		return nil
	}
}
