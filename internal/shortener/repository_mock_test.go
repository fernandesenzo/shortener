package shortener_test

import (
	"context"
	"errors"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
)

type MockRepository struct {
	items            map[string]domain.Link
	shouldError      bool
	collisionCounter int
	tempSaveCalled   bool
	permSaveCalled   bool
}

func (m *MockRepository) TempSave(ctx context.Context, link *domain.TemporaryLink, ttl time.Duration) error {
	m.tempSaveCalled = true
	return m.save(ctx, link)
}

func (m *MockRepository) PermSave(ctx context.Context, link *domain.PermanentLink) error {
	m.permSaveCalled = true
	return m.save(ctx, link)
}

func (m *MockRepository) Get(_ context.Context, code string) (domain.Link, error) {
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

func (m *MockRepository) SetCollisionCounter(amount int) {
	m.collisionCounter = amount
}
func (m *MockRepository) save(_ context.Context, link domain.Link) error {
	if m.shouldError {
		return errors.New("simulated error")
	}
	if m.collisionCounter > 0 {
		m.collisionCounter--
		return shortener.ErrRecordAlreadyExists
	}
	if m.items == nil {
		m.items = make(map[string]domain.Link)
	}
	if m.items[link.GetCode()] != nil {
		return shortener.ErrRecordAlreadyExists
	}
	m.items[link.GetCode()] = link
	return nil
}
