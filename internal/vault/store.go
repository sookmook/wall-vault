package vault

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// defaultServiceList: default service list (auto-seeded if missing from vault.json)
// cloud services: enabled based on key presence → all start as false
// local services: enabled based on connectivity check → all start as false
var defaultServiceList = []*ServiceConfig{
	{ID: "openrouter",     Name: "OpenRouter",        Enabled: false},
	{ID: "google",         Name: "Google Gemini",     Enabled: false},
	{ID: "openai",         Name: "OpenAI",            Enabled: false},
	{ID: "anthropic",      Name: "Anthropic",         Enabled: false},
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
	// migrate: assign SortOrder to existing clients that don't have one
	needsSave := false
	for i, c := range s.clients {
		if c.SortOrder == 0 {
			c.SortOrder = i + 1
			needsSave = true
		}
	}
	sort.Slice(s.clients, func(i, j int) bool {
		return s.clients[i].SortOrder < s.clients[j].SortOrder
	})
	if needsSave {
		defer s.save()
	}
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
		// reorder: sort non-custom services by defaultServiceList order, custom appended last
		order := make(map[string]int, len(defaultServiceList))
		for i, def := range defaultServiceList {
			order[def.ID] = i
		}
		byID := make(map[string]*ServiceConfig, len(s.services))
		for _, sv := range s.services {
			byID[sv.ID] = sv
		}
		sorted := make([]*ServiceConfig, 0, len(s.services))
		for _, def := range defaultServiceList {
			if sv, ok := byID[def.ID]; ok {
				sorted = append(sorted, sv)
			}
		}
		for _, sv := range s.services {
			if sv.Custom {
				sorted = append(sorted, sv)
			}
		}
		s.services = sorted
	}
	// migration: default Enabled=true for existing clients
	needsSave = false
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

	// migrate legacy SHA-256 encrypted keys to Argon2id
	if s.masterPass != "" {
		_ = s.migrateLegacyKeys()
	}

	return nil
}

// migrateLegacyKeys re-encrypts any SHA-256 encrypted key with Argon2id.
// Called once on load. Saves automatically if any key was migrated.
func (s *Store) migrateLegacyKeys() error {
	migrated := 0
	for _, k := range s.keys {
		if !isLegacyEncrypted(k.EncryptedKey) {
			continue
		}
		plain, err := decryptLegacy(k.EncryptedKey, s.masterPass)
		if err != nil {
			continue // skip keys that can't be decrypted (wrong password etc.)
		}
		enc, err := encryptKey(plain, s.masterPass)
		if err != nil {
			continue
		}
		k.EncryptedKey = enc
		migrated++
	}
	if migrated > 0 {
		return s.save()
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

// SetKeyUsage: set absolute usage value reported by proxy heartbeat (idempotent sync).
// Returns true if the value changed.
func (s *Store) SetKeyUsage(id string, tokens int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	for _, k := range s.keys {
		if k.ID == id {
			// auto-reset if it's a new day (guards against stale heartbeat values after midnight)
			if k.UsageDate != today {
				k.TodayUsage = 0
				k.TodayAttempts = 0
				k.UsageDate = today
			}
			if k.TodayUsage != tokens {
				k.TodayUsage = tokens
				_ = s.save()
				return true
			}
			return false
		}
	}
	return false
}

// SetKeyCooldownIfLater: set cooldown only if the new time is later than existing (proxy sync).
// Returns true if changed.
func (s *Store) SetKeyCooldownIfLater(id string, until time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range s.keys {
		if k.ID == id {
			if until.After(k.CooldownUntil) {
				k.CooldownUntil = until
				_ = s.save()
				return true
			}
			return false
		}
	}
	return false
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

// SetKeyAttempts: set total attempt count reported by proxy heartbeat (idempotent sync).
func (s *Store) SetKeyAttempts(id string, attempts int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	for _, k := range s.keys {
		if k.ID == id {
			// auto-reset if it's a new day
			if k.UsageDate != today {
				k.TodayUsage = 0
				k.TodayAttempts = 0
				k.UsageDate = today
			}
			if k.TodayAttempts != attempts {
				k.TodayAttempts = attempts
				_ = s.save()
				return true
			}
			return false
		}
	}
	return false
}

func (s *Store) ResetDailyUsage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	for _, k := range s.keys {
		k.TodayUsage = 0
		k.TodayAttempts = 0
		k.UsageDate = today
	}
	_ = s.save()
}

// BatchUpdateKeyMetrics atomically updates usage, attempts, and cooldowns for all keys
// in a single lock acquisition + save, replacing up to 3N separate save() calls per heartbeat.
func (s *Store) BatchUpdateKeyMetrics(usage, attempts map[string]int, cooldowns map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	changed := false
	for _, k := range s.keys {
		if k.UsageDate != today {
			k.TodayUsage = 0
			k.TodayAttempts = 0
			k.UsageDate = today
			changed = true
		}
		if v, ok := usage[k.ID]; ok && k.TodayUsage != v {
			k.TodayUsage = v
			changed = true
		}
		if v, ok := attempts[k.ID]; ok && k.TodayAttempts != v {
			k.TodayAttempts = v
			changed = true
		}
		if cdStr, ok := cooldowns[k.ID]; ok {
			if until, err := time.Parse(time.RFC3339, cdStr); err == nil {
				if until.After(k.CooldownUntil) {
					k.CooldownUntil = until
					changed = true
				}
			}
		}
	}
	if changed {
		_ = s.save()
	}
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
		if subtle.ConstantTimeCompare([]byte(c.Token), []byte(token)) == 1 {
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
	// assign SortOrder: new client goes to the end
	maxOrder := 0
	for _, existing := range s.clients {
		if existing.SortOrder > maxOrder {
			maxOrder = existing.SortOrder
		}
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
		SortOrder:       maxOrder + 1,
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
			if inp.Avatar != nil {
				c.Avatar = *inp.Avatar
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
			sv.ProxyEnabled = inp.ProxyEnabled
			return s.save()
		}
	}
	// not found — add (custom service)
	clone := *inp
	clone.Custom = true
	s.services = append(s.services, &clone)
	return s.save()
}

// ListProxyEnabledServices: returns service IDs where ProxyEnabled=true
func (s *Store) ListProxyEnabledServices() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var ids []string
	for _, sv := range s.services {
		if sv.ProxyEnabled {
			ids = append(ids, sv.ID)
		}
	}
	return ids
}

// ProxyService: ID + local URL for proxy routing
type ProxyService struct {
	ID       string `json:"id"`
	LocalURL string `json:"local_url,omitempty"`
}

// ListProxyEnabledServicesInfo: returns proxy-enabled services with their local URLs
func (s *Store) ListProxyEnabledServicesInfo() []ProxyService {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []ProxyService
	for _, sv := range s.services {
		if sv.ProxyEnabled {
			result = append(result, ProxyService{ID: sv.ID, LocalURL: sv.LocalURL})
		}
	}
	return result
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

// ReorderClients reorders clients by the given ID list.
// IDs not in the list keep their relative order at the end.
func (s *Store) ReorderClients(order []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	byID := make(map[string]*Client, len(s.clients))
	for _, c := range s.clients {
		byID[c.ID] = c
	}
	seen := make(map[string]bool, len(order))
	reordered := make([]*Client, 0, len(s.clients))
	for i, id := range order {
		if c, ok := byID[id]; ok && !seen[id] {
			c.SortOrder = i + 1
			reordered = append(reordered, c)
			seen[id] = true
		}
	}
	// append any clients not in the order list
	for _, c := range s.clients {
		if !seen[c.ID] {
			c.SortOrder = len(reordered) + 1
			reordered = append(reordered, c)
		}
	}
	s.clients = reordered
	return s.save()
}

// ─── Proxy Status ─────────────────────────────────────────────────────────────

func (s *Store) UpdateProxyStatus(ps *ProxyStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps.UpdatedAt = time.Now()
	if existing, ok := s.proxies[ps.ClientID]; ok && !existing.StartedAt.IsZero() {
		// If the agent was offline (last update too old), reset uptime timer
		if ps.UpdatedAt.Sub(existing.UpdatedAt) > 5*time.Minute {
			ps.StartedAt = ps.UpdatedAt
		} else {
			ps.StartedAt = existing.StartedAt
		}
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
