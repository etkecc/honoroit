package matrix

import (
	"context"
	"time"

	"github.com/sethvargo/go-retry"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

const loginRetry = 15

type accountDataSession struct {
	Token    string
	DeviceID id.DeviceID
}

func (b *Bot) login(username string, password string) error {
	return retry.Fibonacci(context.Background(), loginRetry*time.Second, func(_ context.Context) error {
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
			b.log.Error("cannot authorize using login and password: %v, retrying in %ds...", err, loginRetry)
			return retry.RetryableError(err)
		}
		b.restoreSession()
		b.userID = b.api.UserID
		b.deviceID = b.api.DeviceID

		return nil
	})
}

// restoreSession tries to load previous active session token from account data (if any)
func (b *Bot) restoreSession() {
	b.log.Debug("restoring previous session...")

	var data accountDataSession
	err := b.api.GetAccountData(accountDataSessionToken, &data)
	if err != nil || !b.validateSession(data.Token, data.DeviceID) {
		b.log.Debug("previous session token was not found or invalid: %v", err)
		b.saveSession(b.api.AccessToken, b.api.DeviceID)
		return
	}

	b.log.Debug("previous session restored successfully. Closing current session...")
	if _, logoutErr := b.api.Logout(); logoutErr != nil {
		b.log.Error("cannot logout of current session in favor of previous session: %v", logoutErr)
	}

	b.api.AccessToken = data.Token
	b.api.DeviceID = data.DeviceID
}

func (b *Bot) validateSession(token string, deviceID id.DeviceID) bool {
	valid := true
	// preserve current values
	currentToken := b.api.AccessToken
	currentDeviceID := b.api.DeviceID
	// set new values
	b.api.AccessToken = token
	b.api.DeviceID = deviceID

	if _, err := b.api.GetOwnPresence(); err != nil {
		b.log.Debug("previous session token was not found or invalid: %v", err)
		valid = false
	}

	// restore original values
	b.api.AccessToken = currentToken
	b.api.DeviceID = currentDeviceID
	return valid
}

func (b *Bot) saveSession(token string, deviceID id.DeviceID) {
	data := accountDataSession{
		Token:    token,
		DeviceID: deviceID,
	}

	b.log.Debug("saving session to account data...")
	if err := b.api.SetAccountData(accountDataSessionToken, &data); err != nil {
		b.log.Error("cannot save session to account data: %v", err)
	}
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
