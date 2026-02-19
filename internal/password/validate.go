package password

import "github.com/fernandesenzo/shortener/internal/domain"

func Validate(password string) error {
	if len(password) < 6 {
		return domain.ErrPasswordTooShort
	}
	if len(password) > 20 {
		return domain.ErrPasswordTooLong
	}
	return nil
}
