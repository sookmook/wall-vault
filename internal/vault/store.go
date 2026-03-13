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

// defaultServiceList: default service list (auto-seeded if missing from vault.json)
// cloud services: enabled based on key presence → all start as false
// local services: enabled based on connectivity check → all start as false
var defaultServiceList = []*ServiceConfig{
	{ID: "google",         Name: "Google Gemini",     Enabled: false},
	{ID: "openai",         Name: "OpenAI",            Enabled: false},
	{ID: "anthropic",      Name: "Anthropic",         Enabled: false},
	{ID: "openrouter",     Name: "OpenRouter",        Enabled: false},
	{ID: "github-copilot", Name: "GitHub Copilot",    Enabled: false},
	{ID: "ollama",         Name: "Ollama (Local)",    Enabled: false},
	{ID: "lmstudio",       Name: "LM Studio (Local)", Enabled: false},
	{ID: "vllm",           Name: "vLLM (Local)",      Enabled: false},
}

// Store: thread-safe data store (in-memory + JSON persistence)
type Store struct {
	mu         sync.RWMutex
	keys       []*APIKey
	clients    []*Client
	proxies    map[string]*ProxyStatus
	services   []*ServiceConfig
	settings   StoreSettings
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

// ─── Persistence ─────────────────────────────────────────────────────────────

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
	if snap.Settings != nil {
		s.settings = *snap.Settings
	}
	// load service list + seed default services
	s.services = snap.Services
	if len(s.services) == 0 {
		s.services = make([]*ServiceConfig, len(defaultServiceList))
		copy(s.services, defaultServiceList)
	} else {
		// add missing default services via migration (only Custom=false ones)
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
	// migration: default Enabled=true for existing clients
	needsSave := false
	for _, c := range s.clients {
		if !c.Enabled && c.CreatedAt.Before(time.Now().Add(-time.Second)) {
			// among saved records where Enabled=false, perform migration
			// cannot distinguish "enabled":false explicitly set vs. field missing in JSON
			// → check via raw JSON
			// simplification: set existing clients to enabled=true
		}
	}
	// parse raw JSON to check whether the enabled field is absent
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
	// adjust cloud service enabled state based on key presence
	// (auto-reflected when env var keys are injected or keys already exist)
	if s.reconcileCloudServices() {
		needsSave = true
	}
	if needsSave {
		_ = s.save()
	}
	return nil
}

// ReconcileCloudServices: externally callable version (acquires lock before processing).
func (s *Store) ReconcileCloudServices() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reconcileCloudServices()
}

// reconcileCloudServices: called while lock is already held (inside load() or ReconcileCloudServices).
func (s *Store) reconcileCloudServices() bool {
	keyCounts := map[string]int{}
	for _, k := range s.keys {
		keyCounts[k.Service]++
	}
	changed := false
	for _, sv := range s.services {
		if sv.IsLocal() || sv.Custom {
			continue // do not touch local/custom (decided by user or probe)
		}
		want := keyCounts[sv.ID] > 0
		if sv.Enabled != want {
			sv.Enabled = want
			changed = true
		}
	}
	return changed
}

func (s *Store) save() error {
	proxies := make([]*ProxyStatus, 0, len(s.proxies))
	for _, p := range s.proxies {
		proxies = append(proxies, p)
	}
	settings := s.settings
	snap := storeData{
		Keys:     s.keys,
		Clients:  s.clients,
		Proxies:  proxies,
		Services: s.services,
		Settings: &settings,
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	// atomic write
	tmp := s.dataFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.dataFile)
}

// ─── Key Management ───────────────────────────────────────────────────────────

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

// ─── Client Management ────────────────────────────────────────────────────────

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
			if inp.NewID != nil && *inp.NewID != "" && *inp.NewID != id {
				// check for duplicate ID
				for _, other := range s.clients {
					if other.ID == *inp.NewID {
						return fmt.Errorf("이미 사용 중인 ID: %s", *inp.NewID)
					}
				}
				c.ID = *inp.NewID
			}
			if inp.Name != nil {
				c.Name = *inp.Name
			}
			if inp.Token != nil && *inp.Token != "" {
				c.Token = *inp.Token
			}
			// DefaultService: nil = no change, value present = update (empty value ignored — service-less state is invalid)
			if inp.DefaultService != nil && *inp.DefaultService != "" {
				c.DefaultService = *inp.DefaultService
			}
			// DefaultModel: nil = no change, "" also allowed (means revert to service default model)
			if inp.DefaultModel != nil {
				c.DefaultModel = *inp.DefaultModel
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

// ─── Service Management ───────────────────────────────────────────────────────

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

// UpsertService: update if exists by ID, otherwise add
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
	// not found — add (custom service)
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

// ServiceURLMap: returns service ID → LocalURL map (for models.Registry.Refresh)
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

// ─── Proxy Status ─────────────────────────────────────────────────────────────

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

// ─── UI Settings (theme/language) ────────────────────────────────────────────

func (s *Store) GetSettings() StoreSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *Store) SetTheme(theme string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings.Theme = theme
	return s.save()
}

func (s *Store) SetLang(lang string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings.Lang = lang
	return s.save()
}

// ─── Util ─────────────────────────────────────────────────────────────────────

func newID() string {
	b := make([]byte, 8)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}
