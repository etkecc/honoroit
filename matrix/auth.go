package matrix

import (
	"context"
	"time"

	"github.com/sethvargo/go-retry"
	"maunium.net/go/mautrix"
)

const loginRetry = 15

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
		return nil
	})
}
