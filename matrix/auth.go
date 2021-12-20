package matrix

import (
	"context"
	"time"

	"github.com/sethvargo/go-retry"
	"maunium.net/go/mautrix"
)

const loginRetry = 15

type accountDataSession struct {
	Token string
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
		return nil
	})
}

// restoreSession tries to load previous active session token from account data (if any)
func (b *Bot) restoreSession() {
	var data accountDataSession
	currentToken := b.api.AccessToken
	b.log.Debug("restoring previous session...")

	err := b.api.GetAccountData(accountDataSessionToken, &data)
	if err == nil && data.Token != "" {
		b.log.Debug("previous session token found, trying it...")
		b.api.AccessToken = data.Token
		if _, restoredSessionErr := b.api.GetOwnPresence(); restoredSessionErr == nil {
			b.log.Debug("previous session restored successfully. Closing current session...")
			// we don't need to save current session, because previous one will be used anyway, so - logout!
			b.api.AccessToken = currentToken
			if _, logoutErr := b.api.Logout(); logoutErr != nil {
				b.log.Error("cannot logout of current session in favor of previous session: %v", logoutErr)
			}
			b.api.AccessToken = data.Token
			b.log.Info("restored previous session")
			return
		}

	}

	b.log.Debug("previous session token was not found or invalid")
	b.api.AccessToken = currentToken
	data.Token = currentToken

	b.log.Debug("saving session token to account data...")
	err = b.api.SetAccountData(accountDataSessionToken, &data)
	if err != nil {
		b.log.Error("cannot save session token to account data: %v", err)
	}
}

// hydrate loads auth-related info from already established session
func (b *Bot) hydrate() error {
	b.log.Debug("hydrating bot...")
	whoamiResp, err := b.api.Whoami()
	if err != nil {
		return err
	}
	nameResp, err := b.api.GetOwnDisplayName()
	if err != nil {
		return err
	}

	// following values required for api client to work properly
	b.api.UserID = whoamiResp.UserID
	b.api.DeviceID = whoamiResp.DeviceID

	// AND required for the bot itself to perform some business logic-related stuff
	b.userID = whoamiResp.UserID
	b.deviceID = whoamiResp.DeviceID
	b.name = nameResp.DisplayName

	return nil
}
