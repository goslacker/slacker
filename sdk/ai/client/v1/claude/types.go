package claude

import (
	"encoding/json"
	"github.com/goslacker/slacker/sdk/ai/client/v1"
	"time"
)

type MessageReq struct {
	MaxTokens           int       `json:"max_tokens"`
	Messages            []Message `json:"messages"`
	Model               string    `json:"model"`
	Metadata            *Metadata `json:"metadata,omitempty"`
	StopSequences       []string  `json:"stop_sequences,omitempty"`
	Stream              bool      `json:"stream,omitempty"`
	System              string    `json:"system,omitempty"`
	SystemMultiContents []MultipartContent
	Temperature         *float32    `json:"temperature,omitempty"`
	ToolChoice          *ToolChoice `json:"tool_choice,omitempty"`
	Tools               []Tool      `json:"tools,omitempty"`
	TopK                float32     `json:"top_k,omitempty"`
	TopP                float32     `json:"top_p,omitempty"`
}

func (m *MessageReq) UnmarshalJSON(bytes []byte) (err error) {
	{
		var req struct {
			MaxTokens           int                `json:"max_tokens"`
			Messages            []Message          `json:"messages"`
			Model               string             `json:"model"`
			Metadata            *Metadata          `json:"metadata,omitempty"`
			StopSequences       []string           `json:"stop_sequences,omitempty"`
			Stream              bool               `json:"stream,omitempty"`
			System              string             `json:"system,omitempty"`
			SystemMultiContents []MultipartContent `json:"-"`
			Temperature         *float32           `json:"temperature,omitempty"`
			ToolChoice          *ToolChoice        `json:"tool_choice,omitempty"`
			Tools               []Tool             `json:"tools,omitempty"`
			TopK                float32            `json:"top_k,omitempty"`
			TopP                float32            `json:"top_p,omitempty"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*m = MessageReq(req)
			return nil
		}
	}

	{
		var req struct {
			MaxTokens           int                `json:"max_tokens"`
			Messages            []Message          `json:"messages"`
			Model               string             `json:"model"`
			Metadata            *Metadata          `json:"metadata,omitempty"`
			StopSequences       []string           `json:"stop_sequences,omitempty"`
			Stream              bool               `json:"stream,omitempty"`
			System              string             `json:"-"`
			SystemMultiContents []MultipartContent `json:"system,omitempty"`
			Temperature         *float32           `json:"temperature,omitempty"`
			ToolChoice          *ToolChoice        `json:"tool_choice,omitempty"`
			Tools               []Tool             `json:"tools,omitempty"`
			TopK                float32            `json:"top_k,omitempty"`
			TopP                float32            `json:"top_p,omitempty"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*m = MessageReq(req)
			return nil
		}
	}
	return
}

func (m MessageReq) MarshalJSON() ([]byte, error) {
	if (m.System == "" && len(m.SystemMultiContents) == 0) || m.System != "" {
		req := struct {
			MaxTokens           int                `json:"max_tokens"`
			Messages            []Message          `json:"messages"`
			Model               string             `json:"model"`
			Metadata            *Metadata          `json:"metadata,omitempty"`
			StopSequences       []string           `json:"stop_sequences,omitempty"`
			Stream              bool               `json:"stream,omitempty"`
			System              string             `json:"system,omitempty"`
			SystemMultiContents []MultipartContent `json:"-"`
			Temperature         *float32           `json:"temperature,omitempty"`
			ToolChoice          *ToolChoice        `json:"tool_choice,omitempty"`
			Tools               []Tool             `json:"tools,omitempty"`
			TopK                float32            `json:"top_k,omitempty"`
			TopP                float32            `json:"top_p,omitempty"`
		}(m)
		return json.Marshal(req)
	} else {
		req := struct {
			MaxTokens           int                `json:"max_tokens"`
			Messages            []Message          `json:"messages"`
			Model               string             `json:"model"`
			Metadata            *Metadata          `json:"metadata,omitempty"`
			StopSequences       []string           `json:"stop_sequences,omitempty"`
			Stream              bool               `json:"stream,omitempty"`
			System              string             `json:"-"`
			SystemMultiContents []MultipartContent `json:"system,omitempty"`
			Temperature         *float32           `json:"temperature,omitempty"`
			ToolChoice          *ToolChoice        `json:"tool_choice,omitempty"`
			Tools               []Tool             `json:"tools,omitempty"`
			TopK                float32            `json:"top_k,omitempty"`
			TopP                float32            `json:"top_p,omitempty"`
		}(m)
		return json.Marshal(req)
	}
}

type Metadata struct {
	UserID string `json:"user_id"`
}

type Message struct {
	Role              string `json:"role"`
	Content           string `json:"content"`
	MultipartContents []MultipartContent
}

func (m *Message) UnmarshalJSON(bytes []byte) (err error) {
	{
		var req struct {
			Role              string             `json:"role"`
			Content           string             `json:"content"`
			MultipartContents []MultipartContent `json:"-"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*m = Message(req)
			return nil
		}
	}

	{
		var req struct {
			Role              string             `json:"role"`
			Content           string             `json:"-"`
			MultipartContents []MultipartContent `json:"content"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*m = Message(req)
			return nil
		}
	}
	return
}

func (m Message) MarshalJSON() ([]byte, error) {
	if (m.Content == "" && len(m.MultipartContents) == 0) || m.Content != "" {
		req := struct {
			Role              string             `json:"role"`
			Content           string             `json:"content"`
			MultipartContents []MultipartContent `json:"-"`
		}(m)
		return json.Marshal(req)
	} else {
		req := struct {
			Role              string             `json:"role"`
			Content           string             `json:"-"`
			MultipartContents []MultipartContent `json:"content"`
		}(m)
		return json.Marshal(req)
	}
}

type MultipartContent struct {
	ID                string         `json:"id,omitempty"`
	Input             string         `json:"input,omitempty"`
	Name              string         `json:"name,omitempty"`
	ToolUseID         string         `json:"tool_use_id,omitempty"`
	Type              string         `json:"type"`
	Text              string         `json:"text,omitempty"`
	Source            *SourceContent `json:"source,omitempty"`
	CacheControl      *CacheControl  `json:"cache_control,omitempty"`
	Content           string         `json:"content,omitempty"`
	MultipartContents []MultipartContent
	Citations         []Citation `json:"citations,omitempty"`
	IsError           bool       `json:"is_error,omitempty"`
}

func (m *MultipartContent) UnmarshalJSON(bytes []byte) (err error) {
	{
		var req struct {
			ID                string             `json:"id,omitempty"`
			Input             string             `json:"input,omitempty"`
			Name              string             `json:"name,omitempty"`
			ToolUseID         string             `json:"tool_use_id,omitempty"`
			Type              string             `json:"type"`
			Text              string             `json:"text,omitempty"`
			Source            *SourceContent     `json:"source,omitempty"`
			CacheControl      *CacheControl      `json:"cache_control,omitempty"`
			Content           string             `json:"content,omitempty"`
			MultipartContents []MultipartContent `json:"-"`
			Citations         []Citation         `json:"citations,omitempty"`
			IsError           bool               `json:"is_error,omitempty"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*m = MultipartContent(req)
			return nil
		}
	}

	{
		var req struct {
			ID                string             `json:"id,omitempty"`
			Input             string             `json:"input,omitempty"`
			Name              string             `json:"name,omitempty"`
			ToolUseID         string             `json:"tool_use_id,omitempty"`
			Type              string             `json:"type"`
			Text              string             `json:"text,omitempty"`
			Source            *SourceContent     `json:"source,omitempty"`
			CacheControl      *CacheControl      `json:"cache_control,omitempty"`
			Content           string             `json:"-"`
			MultipartContents []MultipartContent `json:"content,omitempty"`
			Citations         []Citation         `json:"citations,omitempty"`
			IsError           bool               `json:"is_error,omitempty"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*m = MultipartContent(req)
			return nil
		}
	}
	return
}

func (m MultipartContent) MarshalJSON() ([]byte, error) {
	if (m.Content == "" && len(m.MultipartContents) == 0) || m.Content != "" {
		req := struct {
			ID                string             `json:"id,omitempty"`
			Input             string             `json:"input,omitempty"`
			Name              string             `json:"name,omitempty"`
			ToolUseID         string             `json:"tool_use_id,omitempty"`
			Type              string             `json:"type"`
			Text              string             `json:"text,omitempty"`
			Source            *SourceContent     `json:"source,omitempty"`
			CacheControl      *CacheControl      `json:"cache_control,omitempty"`
			Content           string             `json:"content,omitempty"`
			MultipartContents []MultipartContent `json:"-"`
			Citations         []Citation         `json:"citations,omitempty"`
			IsError           bool               `json:"is_error,omitempty"`
		}(m)
		return json.Marshal(req)
	} else {
		req := struct {
			ID                string             `json:"id,omitempty"`
			Input             string             `json:"input,omitempty"`
			Name              string             `json:"name,omitempty"`
			ToolUseID         string             `json:"tool_use_id,omitempty"`
			Type              string             `json:"type"`
			Text              string             `json:"text,omitempty"`
			Source            *SourceContent     `json:"source,omitempty"`
			CacheControl      *CacheControl      `json:"cache_control,omitempty"`
			Content           string             `json:"-"`
			MultipartContents []MultipartContent `json:"content,omitempty"`
			Citations         []Citation         `json:"citations,omitempty"`
			IsError           bool               `json:"is_error,omitempty"`
		}(m)
		return json.Marshal(req)
	}
}

type Citation struct {
	Type            string `json:"type"`
	CitedText       string `json:"cited_text,omitempty"`
	DocumentIndex   int    `json:"document_index,omitempty"`
	DocumentTitle   string `json:"document_title,omitempty"`
	EndPageNumber   int    `json:"end_page_number,omitempty"`
	StartPageNumber int    `json:"start_page_number,omitempty"`
	EndCharIndex    int    `json:"end_char_index,omitempty"`
	StartCharIndex  string `json:"start_char_index,omitempty"`
}

type SourceContent struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type MessageResp struct {
	ID           string             `json:"id"`
	Content      []MultipartContent `json:"content"`
	Model        string             `json:"model"`
	Role         string             `json:"role"`
	StopReason   string             `json:"stop_reason"`
	StopSequence string             `json:"stop_sequence"`
	Type         string             `json:"type"`
	Usage        Usage              `json:"usage"`
	Error        *Error             `json:"error"`
}

type Usage struct {
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}

type ToolChoice struct {
	Type                   string `json:"type"`
	Name                   string `json:"name,omitempty"`
	DisableParallelToolUse *bool  `json:"disable_parallel_tool_use,omitempty"`
}

type Tool struct {
	Name            string        `json:"name"`
	Type            string        `json:"type,omitempty"`
	Description     string        `json:"description,omitempty"`
	InputSchema     *InputSchema  `json:"input_schema,omitempty"`
	CacheControl    *CacheControl `json:"cache_control,omitempty"`
	DisplayHeightPX int           `json:"display_height_px,omitempty"`
	DisplayWidthPX  int           `json:"display_width_px,omitempty"`
	DisplayNumber   int           `json:"display_number,omitempty"`
}

type CacheControl struct {
	Type string `json:"type"`
}

type InputSchema struct {
	Type       string                     `json:"type"`
	Properties map[string]client.Property `json:"properties"`
	Required   []string                   `json:"required,omitempty"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type CreateBatchReq struct {
	Requests []struct {
		CustomerID string     `json:"customer_id"`
		Params     MessageReq `json:"params"`
	} `json:"requests"`
}

type CreateBatchResp struct {
	ArchivedAt        string `json:"archived_at"`
	CancelInitiatedAt string `json:"cancel_initiated_at"`
	CreatedAt         string `json:"created_at"`
	EndedAt           string `json:"ended_at"`
	ExpiresAt         string `json:"expires_at"`
	ID                string `json:"id"`
	ProcessingStatus  string `json:"processing_status"`
	RequestCounts     struct {
		Canceled   int `json:"canceled"`
		Errored    int `json:"errored"`
		Expired    int `json:"expired"`
		Processing int `json:"processing"`
		Succeeded  int `json:"succeeded"`
	} `json:"request_counts"`
	ResultsUrl string `json:"results_url"`
	Type       string `json:"type"`
	Error      *Error `json:"error"`
}

type RetrieveBatchResp struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	ProcessingStatus string `json:"processing_status"`
	RequestCounts    struct {
		Processing int `json:"processing"`
		Succeeded  int `json:"succeeded"`
		Errored    int `json:"errored"`
		Canceled   int `json:"canceled"`
		Expired    int `json:"expired"`
	} `json:"request_counts"`
	EndedAt           time.Time `json:"ended_at"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	ArchivedAt        time.Time `json:"archived_at"`
	CancelInitiatedAt time.Time `json:"cancel_initiated_at"`
	ResultsUrl        string    `json:"results_url"`
	Error             *Error    `json:"error"`
}
