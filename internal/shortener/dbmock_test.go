package shortener_test

import (
	"errors"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
)

type MockRepository struct {
	items       map[string]*domain.Link
	shouldError bool
}

func (m *MockRepository) Save(link *domain.Link) error {
	if m.shouldError {
		return errors.New("simulated error")
	}
	if m.items == nil {
		m.items = make(map[string]*domain.Link)
	}
	m.items[link.Code] = link
	return nil
}

func (m *MockRepository) Get(code string) (*domain.Link, error) {
	_, exists := m.items[code]
	if !exists {
		return nil, shortener.ErrRecordNotFound
	}
	return m.items[code], nil
}
