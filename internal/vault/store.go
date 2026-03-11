package vault

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store: 스레드 안전 데이터 저장소 (메모리 + JSON 영속화)
type Store struct {
	mu         sync.RWMutex
	keys       []*APIKey
	clients    []*Client
	proxies    map[string]*ProxyStatus
	masterPass string
	dataDir    string
	dataFile   string
}

func NewStore(dataDir, masterPass string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("데이터 디렉토리 생성 실패: %w", err)
	}
	s := &Store{
		proxies:    make(map[string]*ProxyStatus),
		masterPass: masterPass,
		dataDir:    dataDir,
		dataFile:   filepath.Join(dataDir, "vault.json"),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("데이터 로드 실패: %w", err)
	}
	return s, nil
}

// ─── 영속화 ──────────────────────────────────────────────────────────────────

func (s *Store) load() error {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return err
	}
	var snap storeData
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	s.keys = snap.Keys
	s.clients = snap.Clients
	if snap.Proxies != nil {
		for _, p := range snap.Proxies {
			s.proxies[p.ClientID] = p
		}
	}
	return nil
}

func (s *Store) save() error {
	proxies := make([]*ProxyStatus, 0, len(s.proxies))
	for _, p := range s.proxies {
		proxies = append(proxies, p)
	}
	snap := storeData{
		Keys:    s.keys,
		Clients: s.clients,
		Proxies: proxies,
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	// 원자적 쓰기
	tmp := s.dataFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.dataFile)
}

// ─── 키 관리 ─────────────────────────────────────────────────────────────────

func (s *Store) ListKeys() []*APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*APIKey, len(s.keys))
	copy(result, s.keys)
	return result
}

func (s *Store) GetAvailableKey(service string) (*APIKey, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, k := range s.keys {
		if k.Service != service {
			continue
		}
		if !k.IsAvailable() {
			continue
		}
		plain, err := decryptKey(k.EncryptedKey, s.masterPass)
		if err != nil {
			continue
		}
		return k, plain, nil
	}
	return nil, "", fmt.Errorf("서비스 '%s' 사용 가능한 키 없음", service)
}

func (s *Store) AddKey(service, plainKey, label string, dailyLimit int) (*APIKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	enc, err := encryptKey(plainKey, s.masterPass)
	if err != nil {
		return nil, err
	}
	k := &APIKey{
		ID:           newID(),
		Service:      service,
		EncryptedKey: enc,
		Label:        label,
		DailyLimit:   dailyLimit,
		CreatedAt:    time.Now(),
	}
	s.keys = append(s.keys, k)
	return k, s.save()
}

func (s *Store) DeleteKey(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, k := range s.keys {
		if k.ID == id {
			s.keys = append(s.keys[:i], s.keys[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("키 없음: %s", id)
}

func (s *Store) RecordKeyUsage(id string, tokens int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range s.keys {
		if k.ID == id {
			k.TodayUsage += tokens
			_ = s.save()
			return
		}
	}
}

func (s *Store) SetKeyCooldown(id string, errCode int, until time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range s.keys {
		if k.ID == id {
			k.CooldownUntil = until
			k.LastError = errCode
			_ = s.save()
			return
		}
	}
}

func (s *Store) ResetDailyUsage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range s.keys {
		k.TodayUsage = 0
	}
	_ = s.save()
}

// ─── 클라이언트 관리 ──────────────────────────────────────────────────────────

func (s *Store) ListClients() []*Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Client, len(s.clients))
	copy(result, s.clients)
	return result
}

func (s *Store) GetClient(id string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.clients {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (s *Store) GetClientByToken(token string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.clients {
		if c.Token == token {
			return c
		}
	}
	return nil
}

func (s *Store) AddClient(id, name, token, service, model string, allowedServices []string) (*Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := &Client{
		ID:              id,
		Name:            name,
		Token:           token,
		DefaultService:  service,
		DefaultModel:    model,
		AllowedServices: allowedServices,
		CreatedAt:       time.Now(),
	}
	s.clients = append(s.clients, c)
	return c, s.save()
}

func (s *Store) UpdateClient(id string, service, model string, allowedServices []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		if c.ID == id {
			if service != "" {
				c.DefaultService = service
			}
			if model != "" {
				c.DefaultModel = model
			}
			if allowedServices != nil {
				c.AllowedServices = allowedServices
			}
			return s.save()
		}
	}
	return fmt.Errorf("클라이언트 없음: %s", id)
}

func (s *Store) DeleteClient(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.clients {
		if c.ID == id {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("클라이언트 없음: %s", id)
}

// ─── 프록시 상태 ──────────────────────────────────────────────────────────────

func (s *Store) UpdateProxyStatus(ps *ProxyStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps.UpdatedAt = time.Now()
	s.proxies[ps.ClientID] = ps
	_ = s.save()
}

func (s *Store) ListProxies() []*ProxyStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*ProxyStatus, 0, len(s.proxies))
	for _, p := range s.proxies {
		result = append(result, p)
	}
	return result
}

// ─── 유틸 ────────────────────────────────────────────────────────────────────

func newID() string {
	b := make([]byte, 8)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}
