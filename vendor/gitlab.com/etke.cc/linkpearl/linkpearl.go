// Package linkpearl represents the library itself
package linkpearl

import (
	"database/sql"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/linkpearl/config"
	"gitlab.com/etke.cc/linkpearl/store"
)

// Linkpearl object
type Linkpearl struct {
	db    *sql.DB
	log   config.Logger
	api   *mautrix.Client
	olm   *crypto.OlmMachine
	store *store.Store
}

// GetClient returns underlying API client
func (l *Linkpearl) GetClient() *mautrix.Client {
	return l.api
}

// GetDB returns underlying DB object
func (l *Linkpearl) GetDB() *sql.DB {
	return l.db
}

// GetStore returns underlying persistent store object, compatible with crypto.Store, crypto.StateStore and mautrix.Storer
func (l *Linkpearl) GetStore() *store.Store {
	return l.store
}

// GetMachine returns underlying OLM machine
func (l *Linkpearl) GetMachine() *crypto.OlmMachine {
	return l.olm
}

// Send a message to the roomID (automatically decide encrypted or not)
func (l *Linkpearl) Send(roomID id.RoomID, content interface{}) (id.EventID, error) {
	if !l.store.IsEncrypted(roomID) {
		resp, err := l.api.SendMessageEvent(roomID, event.EventMessage, content)
		if err != nil {
			return "", err
		}
		return resp.EventID, nil
	}

	encrypted, err := l.olm.EncryptMegolmEvent(roomID, event.EventMessage, content)
	if crypto.IsShareError(err) {
		err = l.olm.ShareGroupSession(roomID, l.store.GetRoomMembers(roomID))
		if err != nil {
			return "", err
		}
		encrypted, err = l.olm.EncryptMegolmEvent(roomID, event.EventMessage, content)
	}

	if err != nil {
		l.log.Error("cannot send encrypted message into %s: %v, sending plaintext...", roomID, err)
		resp, plaintextErr := l.api.SendMessageEvent(roomID, event.EventMessage, content)
		if plaintextErr != nil {
			return "", plaintextErr
		}
		return resp.EventID, nil
	}

	resp, err := l.api.SendMessageEvent(roomID, event.EventEncrypted, encrypted)
	if err != nil {
		return "", err
	}
	return resp.EventID, err
}

// Start performs matrix /sync
func (l *Linkpearl) Start() error {
	l.initSync()
	err := l.api.SetPresence(event.PresenceOnline)
	if err != nil {
		return err
	}
	defer l.Stop()

	l.log.Info("client has been started")
	return l.api.Sync()
}

// Stop the client
func (l *Linkpearl) Stop() {
	l.log.Debug("stopping the client")
	err := l.api.SetPresence(event.PresenceOffline)
	if err != nil {
		l.log.Error("cannot set presence to offile: %v", err)
	}
	l.api.StopSync()
	l.log.Info("client has been stopped")
}
