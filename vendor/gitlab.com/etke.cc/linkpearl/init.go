package linkpearl

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"

	"gitlab.com/etke.cc/linkpearl/config"
	"gitlab.com/etke.cc/linkpearl/store"
)

// New linkpearl
func New(cfg *config.Config) (*Linkpearl, error) {
	api, err := mautrix.NewClient(cfg.Homeserver, "", "")
	if err != nil {
		return nil, err
	}
	api.Logger = cfg.APILogger
	lp := &Linkpearl{
		db:  cfg.DB,
		api: api,
		log: cfg.LPLogger,
	}

	storer := store.New(cfg.DB, cfg.Dialect, cfg.StoreLogger)
	if err = storer.CreateTables(); err != nil {
		return nil, err
	}
	lp.store = storer
	lp.api.Store = storer

	if err = lp.login(cfg.Login, cfg.Password); err != nil {
		return nil, err
	}

	if err = lp.store.WithCrypto(lp.api.UserID, lp.api.DeviceID, cfg.StoreLogger); err != nil {
		return nil, err
	}
	lp.olm = crypto.NewOlmMachine(lp.api, cfg.CryptoLogger, lp.store, lp.store)
	if err = lp.olm.Load(); err != nil {
		return nil, err
	}

	return lp, nil
}
