package chain

import (
	"github.com/goslacker/slacker/ai"
	"sync"
)

func NewHistory() *History {
	return &History{}
}

// History 负责记录聊天历史，**它不会记录第一条system prompt**
type History struct {
	MessageHistory []ai.Message
	historyLock    sync.RWMutex
}

func (c *History) Set(messages ...ai.Message) {
	c.historyLock.Lock()
	defer c.historyLock.Unlock()

	c.MessageHistory = filterFirstSystemPrompt(messages)
}

func (c *History) Get(limit int) (history []ai.Message) {
	start := 0
	if limit > 0 {
		start = len(c.MessageHistory) - 1 - limit
	}

	if start <= 0 {
		return c.MessageHistory
	} else {
		return c.MessageHistory[start:]
	}
}

func filterFirstSystemPrompt(messages []ai.Message) []ai.Message {
	if messages[0].Role == "system" {
		return messages[1:]
	} else {
		return messages
	}
}
