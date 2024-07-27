package chain

import (
	"github.com/goslacker/slacker/ai/client"
	"sync"
)

func NewHistory() *History {
	return &History{}
}

// History 负责记录聊天历史，**它不会记录第一条system prompt**
type History struct {
	MessageHistory []client.Message
	historyLock    sync.RWMutex
}

func (c *History) Set(messages ...client.Message) {
	c.historyLock.Lock()
	defer c.historyLock.Unlock()

	c.MessageHistory = filterFirstSystemPrompt(messages)
}

func (c *History) Get(limit int) (history []client.Message) {
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

func filterFirstSystemPrompt(messages []client.Message) []client.Message {
	if messages[0].Role == "system" {
		return messages[1:]
	} else {
		return messages
	}
}
