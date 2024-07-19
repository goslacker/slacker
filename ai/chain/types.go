package chain

import "github.com/goslacker/slacker/ai"

type ChatNode interface {
	CanGetMessageHistory
	CanSetMessageHistory
}

type CanGetMessageHistory interface {
	SetMessageHistoryGetter(func(limit int) (messages []ai.Message))
}

type CanSetMessageHistory interface {
	SetMessageHistorySetter(func(messages ...ai.Message))
}
