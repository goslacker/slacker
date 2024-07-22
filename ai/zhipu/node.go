package zhipu

import (
	"fmt"
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
	nextID      string
}

func (z ZhipuNode) Run(ctx chain.Context) (nextID string, err error) {
	if z.SaveHistory {
		if c, ok := ctx.(chain.ChatContext); ok {
			z.getMessages = c.GetHistory
			z.setMessages = c.SetHistory
		} else {
			err = fmt.Errorf("required chain.Context to manager history")
			return
		}
	}

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
		Content: ctx.GetParam(z.InKey).(string),
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

	ctx.SetParam(z.GetID(), z.OutKey, resp.Choices[0].Message.Content)
	nextID = z.nextID
	return
}
