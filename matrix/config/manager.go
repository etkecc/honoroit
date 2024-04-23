package config

import (
	"context"
	"strconv"
	"sync"

	"gitlab.com/etke.cc/linkpearl"
)

const (
	configKey     = "cc.etke.honoroit.config"
	mautrix015key = "mautrix015migration"
)

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

	return m
}

func (m *Manager) getConfig(ctx context.Context) map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg, err := m.lp.GetAccountData(ctx, configKey)
	if err == nil {
		m.cfg = cfg
	}

	return m.cfg
}

func (m *Manager) Mautrix015Migration(ctx context.Context) int64 {
	migratedInt, _ := strconv.Atoi(m.getConfig(ctx)[mautrix015key]) //nolint:errcheck // no need
	return int64(migratedInt)
}

// Get config value
func (m *Manager) Get(ctx context.Context, key string) string {
	v := m.getConfig(ctx)[key]
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
func (m *Manager) Set(key, value string) *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cfg[key] = value

	return m
}

// Save config
func (m *Manager) Save(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lp.SetAccountData(ctx, configKey, m.cfg) //nolint:errcheck // we have logs already
}
