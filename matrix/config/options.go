package config

import (
	"strings"

	"gitlab.com/etke.cc/go/mxidwc"
)

var (
	AllowedUsers = &Option{
		Key:         "allow.users",
		Default:     "@*:*",
		Description: "comma-separated list of wildcard rules to allow requests only from specific users",
		Sanitizer: func(s string) string {
			parts := strings.Split(s, ",")
			if len(parts) == 0 {
				return ""
			}

			for i, part := range parts {
				parts[i] = strings.TrimSpace(part)
			}
			_, err := mxidwc.ParsePatterns(parts)
			if err != nil {
				return ""
			}
			return strings.Join(parts, ",")
		},
	}
	IgnoredRooms = &Option{
		Key:         "ignore.rooms",
		Description: "list of room IDs to ignore",
		Sanitizer: func(s string) string {
			s = strings.TrimSpace(s)
			parts := strings.Split(s, ",")
			if len(parts) == 0 {
				return ""
			}

			for i, part := range parts {
				parts[i] = strings.TrimSpace(part)
			}

			return strings.Join(parts, ",")
		},
	}
	IgnoreNoThread = &Option{
		Key:         "ignore.nothread",
		Default:     "false",
		Description: "completely ignores any messages sent outside of thread",
		Sanitizer: func(s string) string {
			s = strings.ToLower(strings.TrimSpace(s))
			if s == "yes" || s == "true" || s == "1" || s == "y" {
				return "true"
			}
			return "false"
		},
	}
	TextPrefixOpen = &Option{
		Key:         "text.prefix.open",
		Default:     "[OPEN]",
		Description: "prefix added to the new thread topics",
		Sanitizer:   strings.TrimSpace,
	}
	TextPrefixDone = &Option{
		Key:         "text.prefix.done",
		Default:     "[DONE]",
		Description: "prefix added to the completed thread topics",
		Sanitizer:   strings.TrimSpace,
	}
	TextGreetings = &Option{
		Key:         "text.greetings",
		Default:     "Thank you for contacting us!",
		Description: "message sent to the customer on the first contact",
		Sanitizer:   strings.TrimSpace,
	}
	TextGreetingsCustomer = &Option{
		Key:         "text.greetings.customer",
		Default:     "Thank you for contacting us! This is your %s request.",
		Description: "message sent to the identified customer on the first contact",
		Sanitizer:   strings.TrimSpace,
	}
	TextJoin = &Option{
		Key:         "text.join",
		Default:     "%s joined the room",
		Description: "message sent to backoffice/threads room when customer joins a room",
		Sanitizer:   strings.TrimSpace,
	}
	TextInvite = &Option{
		Key:         "text.invite",
		Default:     "%s invited %s into the room",
		Description: "message sent to backoffice/threads room when customer invites somebody into a room",
		Sanitizer:   strings.TrimSpace,
	}
	TextLeave = &Option{
		Key:         "text.leave",
		Default:     "%s left the room",
		Description: "message sent to backoffice/threads room when a customer leaves a room",
		Sanitizer:   strings.TrimSpace,
	}
	TextEmptyRoom = &Option{
		Key:         "text.emptyroom",
		Default:     "The last customer left the room.\nConsider that request closed.",
		Description: "message sent to backoffice/threads room when the last customer left a room",
		Sanitizer:   strings.TrimSpace,
	}
	TextError = &Option{
		Key:         "text.error",
		Default:     "Something is wrong. I've notified the developers and they are fixing the issue. Please, try again later or use any other contact method.",
		Description: "message sent to customer if something goes wrong",
		Sanitizer:   strings.TrimSpace,
	}
	TextStart = &Option{
		Key:         "text.start",
		Default:     "The customer was invited to the new room. Send messages into that thread and they will be automatically forwarded.",
		Description: "message that sent into the read as result of the `start` command",
		Sanitizer:   strings.TrimSpace,
	}
	TextCount = &Option{
		Key:         "text.count",
		Default:     "Request has been counted.",
		Description: "message that sent into the read as result of the `count` command",
		Sanitizer:   strings.TrimSpace,
	}
	TextDone = &Option{
		Key:         "text.done",
		Default:     "The operator marked your request as completed. If you think that it's not done yet, please start another 1:1 chat with me to open a new request.",
		Description: "message sent to customer when request marked as done in the threads room",
		Sanitizer:   strings.TrimSpace,
	}

	// Options is full list of the all available options
	Options = ListOfOptions{AllowedUsers, IgnoredRooms, IgnoreNoThread, TextPrefixOpen, TextPrefixDone, TextGreetings, TextGreetingsCustomer, TextJoin, TextInvite, TextLeave, TextEmptyRoom, TextError, TextStart, TextCount, TextDone}
)

type Option struct {
	Key         string
	Default     string
	Description string
	Sanitizer   func(s string) string
}

type ListOfOptions []*Option

func (l ListOfOptions) Find(key string) *Option {
	for _, option := range l {
		if option.Key == key {
			return option
		}
	}
	return nil
}
