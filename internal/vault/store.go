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

// defaultServiceList: 기본 서비스 목록 (vault.json에 없으면 자동 seed)
var defaultServiceList = []*ServiceConfig{
	{ID: "google",         Name: "Google Gemini",     Enabled: true},
	{ID: "openai",         Name: "OpenAI",            Enabled: true},
	{ID: "anthropic",      Name: "Anthropic",         Enabled: true},
	{ID: "openrouter",     Name: "OpenRouter",        Enabled: true},
	{ID: "github-copilot", Name: "GitHub Copilot",    Enabled: true},
	{ID: "ollama",         Name: "Ollama (Local)",    Enabled: true},
	{ID: "lmstudio",       Name: "LM Studio (Local)", Enabled: false},
	{ID: "vllm",           Name: "vLLM (Local)",      Enabled: false},
}

// Store: 스레드 안전 데이터 저장소 (메모리 + JSON 영속화)
type Store struct {
	mu         sync.RWMutex
	keys       []*APIKey
	clients    []*Client
	proxies    map[string]*ProxyStatus
	services   []*ServiceConfig
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
	// 서비스 목록 로드 + 기본 서비스 seed
	s.services = snap.Services
	if len(s.services) == 0 {
		s.services = make([]*ServiceConfig, len(defaultServiceList))
		copy(s.services, defaultServiceList)
	} else {
		// 기본 서비스 중 없는 것을 마이그레이션으로 추가 (Custom=false인 것만)
		existing := make(map[string]bool)
		for _, sv := range s.services {
			existing[sv.ID] = true
		}
		for _, def := range defaultServiceList {
			if !existing[def.ID] {
				clone := *def
				s.services = append(s.services, &clone)
			}
		}
	}
	// 마이그레이션: 기존 클라이언트 Enabled 기본값 true
	needsSave := false
	for _, c := range s.clients {
		if !c.Enabled && c.CreatedAt.Before(time.Now().Add(-time.Second)) {
			// 이미 저장된 레코드 중 Enabled=false인 것은 마이그레이션
			// JSON에 "enabled":false가 명시된 경우와 누락된 경우 구분 불가
			// → raw JSON으로 확인
			// 간단히: 기존 클라이언트는 enabled=true로 설정
		}
	}
	// raw JSON을 파싱해서 enabled 필드 누락 여부 확인
	var rawSnap struct {
		Clients []json.RawMessage `json:"clients"`
	}
	if err2 := json.Unmarshal(data, &rawSnap); err2 == nil {
		for i, raw := range rawSnap.Clients {
			if i >= len(s.clients) {
				break
			}
			var check map[string]interface{}
			if err3 := json.Unmarshal(raw, &check); err3 == nil {
				if _, hasEnabled := check["enabled"]; !hasEnabled {
					s.clients[i].Enabled = true
					needsSave = true
				}
			}
		}
	}
	if snap.Proxies != nil {
		for _, p := range snap.Proxies {
			s.proxies[p.ClientID] = p
		}
	}
	if needsSave {
		_ = s.save()
	}
	return nil
}

func (s *Store) save() error {
	proxies := make([]*ProxyStatus, 0, len(s.proxies))
	for _, p := range s.proxies {
		proxies = append(proxies, p)
	}
	snap := storeData{
		Keys:     s.keys,
		Clients:  s.clients,
		Proxies:  proxies,
		Services: s.services,
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

func (s *Store) AddClient(inp ClientInput) (*Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	enabled := true
	if inp.Enabled != nil {
		enabled = *inp.Enabled
	}
	c := &Client{
		ID:              inp.ID,
		Name:            inp.Name,
		Token:           inp.Token,
		DefaultService:  inp.DefaultService,
		DefaultModel:    inp.DefaultModel,
		AllowedServices: inp.AllowedServices,
		AgentType:       inp.AgentType,
		WorkDir:         inp.WorkDir,
		Description:     inp.Description,
		IPWhitelist:     inp.IPWhitelist,
		Enabled:         enabled,
		CreatedAt:       time.Now(),
	}
	s.clients = append(s.clients, c)
	return c, s.save()
}

func (s *Store) UpdateClient(id string, inp ClientUpdateInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		if c.ID == id {
			if inp.Name != nil {
				c.Name = *inp.Name
			}
			if inp.Token != nil && *inp.Token != "" {
				c.Token = *inp.Token
			}
			if inp.DefaultService != "" {
				c.DefaultService = inp.DefaultService
			}
			if inp.DefaultModel != "" {
				c.DefaultModel = inp.DefaultModel
			}
			if inp.AllowedServices != nil {
				c.AllowedServices = inp.AllowedServices
			}
			if inp.AgentType != nil {
				c.AgentType = *inp.AgentType
			}
			if inp.WorkDir != nil {
				c.WorkDir = *inp.WorkDir
			}
			if inp.Description != nil {
				c.Description = *inp.Description
			}
			if inp.IPWhitelist != nil {
				c.IPWhitelist = inp.IPWhitelist
			}
			if inp.Enabled != nil {
				c.Enabled = *inp.Enabled
			}
			return s.save()
		}
	}
	return fmt.Errorf("클라이언트 없음: %s", id)
}

// ─── 서비스 관리 ──────────────────────────────────────────────────────────────

func (s *Store) ListServices() []*ServiceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*ServiceConfig, len(s.services))
	copy(result, s.services)
	return result
}

func (s *Store) GetService(id string) *ServiceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sv := range s.services {
		if sv.ID == id {
			return sv
		}
	}
	return nil
}

// UpsertService: ID 기준으로 있으면 업데이트, 없으면 추가
func (s *Store) UpsertService(inp *ServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sv := range s.services {
		if sv.ID == inp.ID {
			if inp.Name != "" {
				sv.Name = inp.Name
			}
			sv.LocalURL = inp.LocalURL
			sv.Enabled = inp.Enabled
			return s.save()
		}
	}
	// 없으면 추가 (커스텀 서비스)
	clone := *inp
	clone.Custom = true
	s.services = append(s.services, &clone)
	return s.save()
}

func (s *Store) DeleteService(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sv := range s.services {
		if sv.ID == id {
			if !sv.Custom {
				return fmt.Errorf("기본 서비스는 삭제할 수 없습니다: %s", id)
			}
			s.services = append(s.services[:i], s.services[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("서비스 없음: %s", id)
}

// ServiceURLMap: 서비스 ID → LocalURL 맵 반환 (models.Registry.Refresh용)
func (s *Store) ServiceURLMap() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m := make(map[string]string, len(s.services))
	for _, sv := range s.services {
		if sv.LocalURL != "" {
			m[sv.ID] = sv.LocalURL
		}
	}
	return m
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
	if existing, ok := s.proxies[ps.ClientID]; ok && !existing.StartedAt.IsZero() {
		ps.StartedAt = existing.StartedAt
	} else if ps.StartedAt.IsZero() {
		ps.StartedAt = ps.UpdatedAt
	}
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
