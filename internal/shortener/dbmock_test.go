package shortener_test

import (
	"errors"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type MockRepository struct {
	items       map[string]*domain.Link
	shouldError bool
}

func (m *MockRepository) Save(link *domain.Link) error {
	if m.shouldError {
		return errors.New("simulated error")
	}
	if m.items[link.Code] == nil {
		m.items = make(map[string]*domain.Link)
	}
	m.items[link.Code] = link
	return nil
}

func (m *MockRepository) Get(code string) (*domain.Link, error) {
	return nil, nil
}
