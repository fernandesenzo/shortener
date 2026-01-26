package shortener_test

import (
	"context"
	"errors"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
)

type MockRepository struct {
	items       map[string]*domain.Link
	shouldError bool
}

func (m *MockRepository) Save(_ context.Context, link *domain.Link) error {
	if m.shouldError {
		return errors.New("simulated error")
	}
	if m.items == nil {
		m.items = make(map[string]*domain.Link)
	}
	m.items[link.Code] = link
	return nil
}

func (m *MockRepository) Get(_ context.Context, code string) (*domain.Link, error) {
	if m.shouldError {
		return nil, errors.New("simulated error")
	}
	_, exists := m.items[code]
	if !exists {
		return nil, shortener.ErrRecordNotFound
	}
	return m.items[code], nil
}

func (m *MockRepository) SetShouldError(shouldError bool) {
	m.shouldError = shouldError
}
