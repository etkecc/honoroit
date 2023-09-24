package config

import (
	"sync"

	"gitlab.com/etke.cc/linkpearl"
)

const configKey = "cc.etke.honoroit.config"

// Manager of configs
type Manager struct {
	mu  *sync.Mutex
	cfg map[string]string
	lp  *linkpearl.Linkpearl
}

// New config manager
func New(lp *linkpearl.Linkpearl) *Manager {
	m := &Manager{
		mu:  &sync.Mutex{},
		cfg: make(map[string]string),
		lp:  lp,
	}
	m.migrate()

	return m
}

func (m *Manager) getConfig() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg, err := m.lp.GetAccountData(configKey)
	if err == nil {
		m.cfg = cfg
	}

	return m.cfg
}

// Get config value
func (m *Manager) Get(key string) string {
	v := m.getConfig()[key]
	if v != "" {
		return v
	}
	o := Options.Find(key)
	if o != nil {
		return o.Default
	}
	return ""
}

// Set config value without saving
func (m *Manager) Set(key, value string) *Manager { //nolint:unparam // lies
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cfg[key] = value

	return m
}

// Save config
func (m *Manager) Save() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lp.SetAccountData(configKey, m.cfg) //nolint:errcheck // we have logs already
}
