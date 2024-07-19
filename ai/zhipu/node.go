package zhipu

import (
	"github.com/goslacker/slacker/ai"
	"github.com/goslacker/slacker/ai/chain"
)

type ZhipuNode struct {
	chain.NodeInfo
	InKey        string
	OutKey       string
	SystemPrompt string
	SaveHistory  bool
	Model        string
	Limit        int
	*Client
	getMessages func(limit int) (messages []ai.Message)
	setMessages func(message ...ai.Message)
}

func (z *ZhipuNode) SetMessageHistoryGetter(f func(limit int) (messages []ai.Message)) {
	z.getMessages = f
}

func (z *ZhipuNode) SetMessageHistorySetter(f func(message ...ai.Message)) {
	z.setMessages = f
}

func (z ZhipuNode) Run(params ai.Params) (nextID string, err error) {
	var history []ai.Message
	if z.SaveHistory {
		history = z.getMessages(z.Limit)
	}

	if len(history) == 0 {
		history = append(history, ai.Message{
			Role:    "system",
			Content: z.SystemPrompt,
		})
	}
	history = append(history, ai.Message{
		Role:    "user",
		Content: params.Get(z.InKey).(string),
	})
	messages, err := MessagesFromStandard(history...)
	if err != nil {
		return
	}
	resp, err := z.Client.ChatCompletion(&ChatCompletionReq{
		Model:    z.Model,
		Messages: messages,
	})
	if err != nil {
		return
	}
	if z.SaveHistory {
		var stdMessage []ai.Message
		stdMessage, err = ToStandardMessages(*resp.Choices[0].Message)
		if err != nil {
			return
		}
		history = append(history, stdMessage...)
		z.setMessages(history...)
	}

	params.Set(z.GetID(), z.OutKey, resp.Choices[0].Message.Content)
	nextID = z.NextID
	return
}
