package memory

import (
	"context"
	"sync"
	"time"

	"github.com/zalgonoise/x/politebot/internal/repository"
)

type Clock interface {
	Now() time.Time
}

type InMemory struct {
	mu                *sync.RWMutex
	userAngyPointsMap map[string]Entry

	regmu       *sync.RWMutex
	registryMap map[string]struct{}

	clock Clock
}

type Entry struct {
	angyPoints int
	lastUpdate time.Time
}

func (m *InMemory) GetAngyPoints(_ context.Context, user string) (int, time.Time, error) {
	m.mu.RLock()
	entry, ok := m.userAngyPointsMap[user]
	m.mu.RUnlock()

	if !ok {
		return 0, time.Time{}, nil
	}

	return entry.angyPoints, entry.lastUpdate, nil
}

func (m *InMemory) ListAngyPoints(_ context.Context) (map[string]int, error) {
	m.mu.RLock()
	if len(m.userAngyPointsMap) == 0 {
		m.mu.RUnlock()

		return nil, repository.ErrNotFound
	}

	dst := make(map[string]int, len(m.userAngyPointsMap))
	for k, v := range m.userAngyPointsMap {
		dst[k] = v.angyPoints
	}

	m.mu.RUnlock()

	return dst, nil
}

func (m *InMemory) AddAngyPoints(_ context.Context, user string, n int) (int, error) {
	m.mu.Lock()
	current := m.userAngyPointsMap[user]
	entry := Entry{
		angyPoints: current.angyPoints + n,
		lastUpdate: m.clock.Now(),
	}
	m.userAngyPointsMap[user] = entry
	m.mu.Unlock()

	return entry.angyPoints, nil
}

func (m *InMemory) RegisterUser(_ context.Context, userID string) error {
	m.regmu.Lock()
	_, ok := m.registryMap[userID]

	if !ok {
		m.registryMap[userID] = struct{}{}
	}

	m.regmu.Unlock()

	return nil
}

func NewInMemory(clock Clock) *InMemory {
	return &InMemory{
		userAngyPointsMap: make(map[string]Entry, 64),
		mu:                new(sync.RWMutex),
		clock:             clock,
	}
}
