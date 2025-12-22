package memory

import (
	"context"
	"sync"
	"time"

	"github.com/zalgonoise/x/cookies/internal/repository"
)

type Clock interface {
	Now() time.Time
}

type InMemory struct {
	mu             *sync.RWMutex
	userCookiesMap map[string]Entry

	clock Clock
}

type Entry struct {
	cookies    int
	lastUpdate time.Time
}

func (m *InMemory) GetCookies(ctx context.Context, user string) (int, time.Time, error) {
	m.mu.RLock()
	entry, ok := m.userCookiesMap[user]
	m.mu.RUnlock()

	if !ok {
		return 0, time.Time{}, nil
	}

	return entry.cookies, entry.lastUpdate, nil
}

func (m *InMemory) ListCookies(ctx context.Context) (map[string]int, error) {
	m.mu.RLock()
	if len(m.userCookiesMap) == 0 {
		m.mu.RUnlock()

		return nil, repository.ErrNotFound
	}

	dst := make(map[string]int, len(m.userCookiesMap))
	for k, v := range m.userCookiesMap {
		dst[k] = v.cookies
	}

	m.mu.RUnlock()

	return dst, nil
}

func (m *InMemory) AddCookie(ctx context.Context, user string, n int) (int, error) {
	m.mu.Lock()
	current := m.userCookiesMap[user]
	entry := Entry{
		cookies:    current.cookies + n,
		lastUpdate: m.clock.Now(),
	}
	m.userCookiesMap[user] = entry
	m.mu.Unlock()

	return entry.cookies, nil
}

func (m *InMemory) SwapCookies(ctx context.Context, from, to string, n int) (int, int, error) {
	m.mu.Lock()
	entryFrom, ok := m.userCookiesMap[from]
	if !ok {
		m.mu.Unlock()
		return 0, 0, repository.ErrNotFound
	}
	entryTo, ok := m.userCookiesMap[to]
	if !ok {
		m.userCookiesMap[to] = Entry{
			cookies:    n,
			lastUpdate: m.clock.Now(),
		}

		requesterCurrent := entryFrom.cookies - n
		m.userCookiesMap[from] = Entry{
			cookies:    requesterCurrent,
			lastUpdate: m.clock.Now(),
		}
		m.mu.Unlock()

		return requesterCurrent, n, nil
	}

	targetCurrent := entryTo.cookies + n
	m.userCookiesMap[to] = Entry{
		cookies:    targetCurrent,
		lastUpdate: entryTo.lastUpdate,
	}

	requesterCurrent := entryFrom.cookies - n
	m.userCookiesMap[from] = Entry{
		cookies:    requesterCurrent,
		lastUpdate: m.clock.Now(),
	}
	m.mu.Unlock()

	return requesterCurrent, targetCurrent, nil
}

func (m *InMemory) EatCookie(ctx context.Context, user string) (int, error) {
	m.mu.Lock()
	entry, ok := m.userCookiesMap[user]
	if !ok {
		m.mu.Unlock()

		return 0, repository.ErrNotFound
	}
	current := entry.cookies - 1
	m.userCookiesMap[user] = Entry{
		cookies:    current,
		lastUpdate: entry.lastUpdate,
	}

	m.mu.Unlock()

	return current, nil
}

func NewInMemory(clock Clock) *InMemory {
	return &InMemory{
		userCookiesMap: make(map[string]Entry, 64),
		mu:             new(sync.RWMutex),
		clock:          clock,
	}
}
