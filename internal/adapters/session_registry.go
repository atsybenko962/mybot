package adapters

import (
	"context"
	"encoding/json"
	lru "github.com/hashicorp/golang-lru/v2"
	"log"
	"os"
	"sync"
	"time"
)

const sessionsFilePath = "sessions.json"

type ChatSession struct {
	ChatID     int64     `json:"chat_id"`
	Title      string    `json:"title"`
	Type       string    `json:"type"` // "group", "supergroup"
	JoinedAt   time.Time `json:"joined_at"`
	LastActive time.Time `json:"last_active"`
	IsRemoved  bool      `json:"is_removed"`
}

type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[string]*ChatSession          // key → session
	lru      *lru.Cache[string, *ChatSession] // или просто TTL-map
	cancels  map[string]context.CancelFunc    // key → cancel воркера
}

func NewSessionRegistry(maxEntries int) *SessionRegistry {
	cache, _ := lru.New[string, *ChatSession](maxEntries)
	return &SessionRegistry{
		sessions: make(map[string]*ChatSession),
		lru:      cache,
		cancels:  make(map[string]context.CancelFunc),
	}
}

// Upsert — добавляет или обновляет сессию (при добавлении в чат)
func (r *SessionRegistry) Upsert(key string, session *ChatSession) {
	r.mu.Lock()
	r.sessions[key] = session
	r.lru.Add(key, session) // для LRU-ограничения по памяти
	r.mu.Unlock()

	r.saveToFile()
}

// AttachWorkerCancel регистрирует cancel-функцию воркера для чата.
// Если для этого ключа уже есть воркер — он будет остановлен перед заменой.
func (r *SessionRegistry) AttachWorkerCancel(key string, cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if old, ok := r.cancels[key]; ok && old != nil {
		old()
	}
	r.cancels[key] = cancel
}

// stopWorkerLocked останавливает воркер без выхода за пределы мьютекса.
func (r *SessionRegistry) stopWorkerLocked(key string) {
	if cancel, ok := r.cancels[key]; ok && cancel != nil {
		cancel()
		delete(r.cancels, key)
	}
}

// Remove — помечает сессию как удалённую (мягкое удаление) и останавливает воркер
func (r *SessionRegistry) Remove(key string) {
	r.mu.Lock()
	if s, ok := r.sessions[key]; ok {
		s.IsRemoved = true
		s.LastActive = time.Now()
	}
	r.stopWorkerLocked(key)
	r.mu.Unlock()

	r.saveToFile()
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

// GetAllActive — для метрик/мониторинга и рестарта воркеров
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

// LoadFromFile восстанавливает сессии из файла при старте сервиса.
func (r *SessionRegistry) LoadFromFile() error {
	data, err := os.ReadFile(sessionsFilePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var wrapper struct {
		Sessions map[string]*ChatSession `json:"sessions"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for k, s := range wrapper.Sessions {
		r.sessions[k] = s
	}

	return nil
}

// saveToFile сохраняет текущее состояние сессий на диск (без goroutine-канцелов).
func (r *SessionRegistry) saveToFile() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wrapper := struct {
		Sessions map[string]*ChatSession `json:"sessions"`
	}{
		Sessions: r.sessions,
	}

	data, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		log.Printf("⚠️ Не удалось сериализовать сессии: %v", err)
		return
	}

	if err := os.WriteFile(sessionsFilePath, data, 0o644); err != nil {
		log.Printf("⚠️ Не удалось сохранить сессии в файл: %v", err)
	}
}
