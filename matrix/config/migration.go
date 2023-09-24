package config

import (
	"strconv"
	"strings"
	"time"

	"gitlab.com/etke.cc/go/env"
)

// TODO: remove whole file
const mautrix015key = "mautrix015migration"

// TODO: remove whole file
var oldDefaults = map[string]string{
	"text.prefix.open": "[OPEN]",
	"text.prefix.done": "[DONE]",
	"text.greetings": `Hello,
Your message was sent to operators.
Please, keep calm and wait for the answer (usually, it takes 1-2 days).
Don't forget that instant messenger is the same communication channel as email, so don't expect an instant response.
Please be advised that requests that are older than 7 days are eligible for automatic removal on our end.`,
	"text.join":   "New customer (%s) joined the room",
	"text.invite": "Customer (%s) invited another user (%s) into the room",
	"text.leave":  "Customer (%s) left the room",
	"text.error": `Something is wrong.
I notified the developers and they are fixing the issue.
Please, try again later or use any other contact method.`,
	"text.emptyroom": "The last customer left the room.\nConsider that request closed.",
	"text.start":     "The customer was invited to the new room. Send messages into that thread and they will be automatically forwarded.",
	"text.count":     "Request has been counted.",
	"text.done": `The operator marked your request as completed.
If you think that it's not done yet, please start another 1:1 chat with me to open a new request.`,
	"text.noencryption": `Unfortunately, encryption is disabled to prevent common decryption issues among customers.
Please, start a new un-encrypted chat with me.`,
}

// TODO remove after some time
func (m *Manager) migrate() {
	var changed bool
	current := m.getConfig()

	if _, ok := current[mautrix015key]; !ok {
		m.Set(mautrix015key, strconv.Itoa(int(time.Now().UTC().UnixMilli())))
		changed = true
	}

	env.SetPrefix("honoroit")
	for key, oldDefault := range oldDefaults {
		if _, ok := current[key]; ok {
			continue
		}
		currentEnv := env.String(key, oldDefault)
		m.Set(key, currentEnv)
		changed = true
	}

	if _, ok := current["allow.users"]; !ok {
		currentEnv := env.Slice("allowedusers")
		if len(currentEnv) != 0 {
			m.Set("allow.users", strings.Join(currentEnv, ","))
			changed = true
		}
	}

	if _, ok := current["ignore.rooms"]; !ok {
		currentEnv := env.Slice("ignoredrooms")
		if len(currentEnv) != 0 {
			m.Set("ignore.rooms", strings.Join(currentEnv, ","))
			changed = true
		}
	}

	if _, ok := current["ignore.nothread"]; !ok {
		currentEnv := env.Bool("ignorenothread")
		if currentEnv {
			m.Set("ignore.nothread", "true")
		} else {
			m.Set("ignore.nothread", "false")
		}
		changed = true
	}

	if changed {
		m.Save()
	}
}

// TODO remove after some time
func (m *Manager) Mautrix015Migration() int64 {
	migratedInt, _ := strconv.Atoi(m.getConfig()[mautrix015key]) //nolint:errcheck // no need
	return int64(migratedInt)
}
