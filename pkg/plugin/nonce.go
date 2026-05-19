package plugin

import (
	"sync"
	"time"
)

type MemoryNonceStore struct {
	mu     sync.Mutex
	nonces map[string]time.Time
}

func NewMemoryNonceStore() *MemoryNonceStore {
	return &MemoryNonceStore{nonces: map[string]time.Time{}}
}

func (s *MemoryNonceStore) UseNonce(pluginKey string, nonce string, expiresAt time.Time, now time.Time) error {
	if s == nil {
		return nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.nonces == nil {
		s.nonces = map[string]time.Time{}
	}
	for key, expiry := range s.nonces {
		if expiry.Before(now) {
			delete(s.nonces, key)
		}
	}
	key := pluginKey + ":" + nonce
	if _, ok := s.nonces[key]; ok {
		return ErrNonceReused
	}
	s.nonces[key] = expiresAt
	return nil
}
