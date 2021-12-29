package matrix

import (
	"errors"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

func (b *Bot) login(username string, password string) error {
	if err := b.restoreSession(); err == nil {
		b.log.Debug("session restored successfully")
		return nil
	}

	b.log.Debug("auth using login and password...")
	_, err := b.api.Login(&mautrix.ReqLogin{
		Type: "m.login.password",
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: username,
		},
		Password:         password,
		StoreCredentials: true,
	})
	if err != nil {
		b.log.Error("cannot authorize using login and password: %v", err)
		return err
	}
	b.store.SaveSession(b.api.UserID, b.api.DeviceID, b.api.AccessToken)

	return nil
}

// restoreSession tries to load previous active session token from db (if any)
func (b *Bot) restoreSession() error {
	b.log.Debug("restoring previous session...")

	userID, deviceID, token := b.store.LoadSession()
	if userID == "" || deviceID == "" || token == "" {
		return errors.New("cannot restore session from db")
	}
	if !b.validateSession(userID, deviceID, token) {
		return errors.New("restored session is invalid")
	}

	b.api.AccessToken = token
	b.api.UserID = userID
	b.api.DeviceID = deviceID
	return nil
}

func (b *Bot) validateSession(userID id.UserID, deviceID id.DeviceID, token string) bool {
	valid := true
	// preserve current values
	currentToken := b.api.AccessToken
	currentUserID := b.api.UserID
	currentDeviceID := b.api.DeviceID
	// set new values
	b.api.AccessToken = token
	b.api.UserID = userID
	b.api.DeviceID = deviceID

	if _, err := b.api.GetOwnPresence(); err != nil {
		b.log.Debug("previous session token was not found or invalid: %v", err)
		valid = false
	}

	// restore original values
	b.api.AccessToken = currentToken
	b.api.UserID = currentUserID
	b.api.DeviceID = currentDeviceID
	return valid
}

// hydrate loads auth-related info from already established session
func (b *Bot) hydrate() error {
	b.log.Debug("hydrating bot...")
	nameResp, err := b.api.GetOwnDisplayName()
	if err != nil {
		return err
	}
	b.name = nameResp.DisplayName

	return nil
}
