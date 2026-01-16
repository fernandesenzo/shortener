package shortener

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
)

// Sentinel Errors
var (
	ErrGenCode      = errors.New("error generating code")
	ErrLinkNotSaved = errors.New("link not saved")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Shorten(originalURL string) (*domain.Link, error) {
	code, err := generateCode(6)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenCode, err)
	}
	link := &domain.Link{
		Code:        code,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}
	if err := s.repo.Save(link); err != nil {

		return nil, fmt.Errorf("%w: %v", ErrLinkNotSaved, err)
	}
	return link, nil
}

// private functions

const validChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateCode(n int) (string, error) {
	code := make([]byte, n)
	maxValue := big.NewInt(int64(len(validChars)))

	for i := range code {
		num, err := rand.Int(rand.Reader, maxValue)
		if err != nil {
			return "", err
		}
		code[i] = validChars[num.Int64()]
	}
	return string(code), nil
}
