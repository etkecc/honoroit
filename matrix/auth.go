package matrix

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-retry"
	"maunium.net/go/mautrix"
)

func (b *Bot) login(username string, password string) error {
	return retry.Fibonacci(context.Background(), 1*time.Second, func(_ context.Context) error {
		fmt.Println("logging in...")
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
			fmt.Println("login error:", err)
			return retry.RetryableError(err)
		}
		return nil
	})
}
