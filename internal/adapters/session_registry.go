package adapters

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"sync"
	"time"
)

type ChatSession struct {
	ChatID     int64
	Title      string
	Type       string // "group", "supergroup"
	JoinedAt   time.Time
	LastActive time.Time
	IsRemoved  bool
}

type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[string]*ChatSession          // key → session
	lru      *lru.Cache[string, *ChatSession] // или просто TTL-map
}

func NewSessionRegistry(maxEntries int) *SessionRegistry {
	cache, _ := lru.New[string, *ChatSession](maxEntries)
	return &SessionRegistry{
		sessions: make(map[string]*ChatSession),
		lru:      cache,
	}
}

// Upsert — добавляет или обновляет сессию (при добавлении в чат)
func (r *SessionRegistry) Upsert(key string, session *ChatSession) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[key] = session
	r.lru.Add(key, session) // для LRU-ограничения по памяти
}

// Remove — помечает сессию как удалённую (мягкое удаление)
func (r *SessionRegistry) Remove(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if s, ok := r.sessions[key]; ok {
		s.IsRemoved = true
		s.LastActive = time.Now()
	}
}

// GetActive — возвращает активную сессию (не удалённую)
func (r *SessionRegistry) GetActive(key string) (*ChatSession, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.sessions[key]
	if !ok || s.IsRemoved {
		return nil, false
	}
	s.LastActive = time.Now() // обновляем активность
	return s, true
}

// GetAllActive — для метрик/мониторинга
func (r *SessionRegistry) GetAllActive() map[string]*ChatSession {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make(map[string]*ChatSession)
	for k, s := range r.sessions {
		if !s.IsRemoved {
			res[k] = s
		}
	}
	return res
}
