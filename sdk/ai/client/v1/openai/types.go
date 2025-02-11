package openai

import (
	"encoding/json"
	"github.com/goslacker/slacker/core/tool"
	"github.com/goslacker/slacker/sdk/ai/client"
)

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

func FromStdChatCompletionReq(req *client.ChatCompletionReq) (r *ChatCompletionReq) {
	tool.SimpleMap(&r, req)
	return
}

type ChatCompletionReq struct {
	Model               string                  `json:"model"`
	Messages            []Message               `json:"messages"`
	Store               bool                    `json:"store,omitempty"`
	ReasoningEffort     string                  `json:"reasoning_effort,omitempty"`
	Metadata            map[string]string       `json:"metadata,omitempty"`
	FrequencyPenalty    float32                 `json:"frequency_penalty,omitempty"`
	LogitBias           map[string]int          `json:"logit_bias,omitempty"`
	Logprobs            bool                    `json:"logprobs,omitempty"`
	TopLogprobs         *int                    `json:"top_logprobs,omitempty"`
	MaxCompletionTokens int                     `json:"max_completion_tokens,omitempty"`
	N                   int                     `json:"n,omitempty"`
	Modalities          []string                `json:"modalities,omitempty"`
	Prediction          *Prediction             `json:"prediction,omitempty"`
	Audio               *ChatCompletionReqAudio `json:"audio,omitempty"`
	PresencePenalty     int                     `json:"presence_penalty,omitempty"`
	ResponseFormat      *ResponseFormat         `json:"response_format,omitempty"`
	Seed                int                     `json:"seed,omitempty"`
	ServiceTier         string                  `json:"service_tier,omitempty"`
	Stop                []string                `json:"stop,omitempty"`
	Stream              bool                    `json:"stream,omitempty"`
	StreamOptions       *StreamOptions          `json:"stream_options,omitempty"`
	Temperature         *float32                `json:"temperature,omitempty"`
	TopP                *float32                `json:"top_p,omitempty"`
	Tools               []Tool                  `json:"tools,omitempty"`
	ToolChoice          string                  `json:"tool_choice,omitempty"`
	ToolChoiceObj       *ToolChoice
	ParallelToolCalls   *bool  `json:"parallel_tool_calls,omitempty"`
	User                string `json:"user,omitempty"`
}

func (c *ChatCompletionReq) UnmarshalJSON(bytes []byte) (err error) {
	{
		var req struct {
			Model               string                  `json:"model"`
			Messages            []Message               `json:"messages"`
			Store               bool                    `json:"store,omitempty"`
			ReasoningEffort     string                  `json:"reasoning_effort,omitempty"`
			Metadata            map[string]string       `json:"metadata,omitempty"`
			FrequencyPenalty    float32                 `json:"frequency_penalty,omitempty"`
			LogitBias           map[string]int          `json:"logit_bias,omitempty"`
			Logprobs            bool                    `json:"logprobs,omitempty"`
			TopLogprobs         *int                    `json:"top_logprobs,omitempty"`
			MaxCompletionTokens int                     `json:"max_completion_tokens,omitempty"`
			N                   int                     `json:"n,omitempty"`
			Modalities          []string                `json:"modalities,omitempty"`
			Prediction          *Prediction             `json:"prediction,omitempty"`
			Audio               *ChatCompletionReqAudio `json:"audio,omitempty"`
			PresencePenalty     int                     `json:"presence_penalty,omitempty"`
			ResponseFormat      *ResponseFormat         `json:"response_format,omitempty"`
			Seed                int                     `json:"seed,omitempty"`
			ServiceTier         string                  `json:"service_tier,omitempty"`
			Stop                []string                `json:"stop,omitempty"`
			Stream              bool                    `json:"stream,omitempty"`
			StreamOptions       *StreamOptions          `json:"stream_options,omitempty"`
			Temperature         *float32                `json:"temperature,omitempty"`
			TopP                *float32                `json:"top_p,omitempty"`
			Tools               []Tool                  `json:"tools,omitempty"`
			ToolChoice          string                  `json:"tool_choice,omitempty"`
			ToolChoiceObj       *ToolChoice             `json:"-"`
			ParallelToolCalls   *bool                   `json:"parallel_tool_calls,omitempty"`
			User                string                  `json:"user,omitempty"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*c = ChatCompletionReq(req)
			return nil
		}
	}

	{
		var req struct {
			Model               string                  `json:"model"`
			Messages            []Message               `json:"messages"`
			Store               bool                    `json:"store,omitempty"`
			ReasoningEffort     string                  `json:"reasoning_effort,omitempty"`
			Metadata            map[string]string       `json:"metadata,omitempty"`
			FrequencyPenalty    float32                 `json:"frequency_penalty,omitempty"`
			LogitBias           map[string]int          `json:"logit_bias,omitempty"`
			Logprobs            bool                    `json:"logprobs,omitempty"`
			TopLogprobs         *int                    `json:"top_logprobs,omitempty"`
			MaxCompletionTokens int                     `json:"max_completion_tokens,omitempty"`
			N                   int                     `json:"n,omitempty"`
			Modalities          []string                `json:"modalities,omitempty"`
			Prediction          *Prediction             `json:"prediction,omitempty"`
			Audio               *ChatCompletionReqAudio `json:"audio,omitempty"`
			PresencePenalty     int                     `json:"presence_penalty,omitempty"`
			ResponseFormat      *ResponseFormat         `json:"response_format,omitempty"`
			Seed                int                     `json:"seed,omitempty"`
			ServiceTier         string                  `json:"service_tier,omitempty"`
			Stop                []string                `json:"stop,omitempty"`
			Stream              bool                    `json:"stream,omitempty"`
			StreamOptions       *StreamOptions          `json:"stream_options,omitempty"`
			Temperature         *float32                `json:"temperature,omitempty"`
			TopP                *float32                `json:"top_p,omitempty"`
			Tools               []Tool                  `json:"tools,omitempty"`
			ToolChoice          string                  `json:"-"`
			ToolChoiceObj       *ToolChoice             `json:"tool_choice,omitempty"`
			ParallelToolCalls   *bool                   `json:"parallel_tool_calls,omitempty"`
			User                string                  `json:"user,omitempty"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*c = ChatCompletionReq(req)
			return nil
		}
	}
	return
}

func (c ChatCompletionReq) MarshalJSON() ([]byte, error) {
	if (c.ToolChoice == "" && c.ToolChoiceObj == nil) || c.ToolChoice != "" {
		req := struct {
			Model               string                  `json:"model"`
			Messages            []Message               `json:"messages"`
			Store               bool                    `json:"store,omitempty"`
			ReasoningEffort     string                  `json:"reasoning_effort,omitempty"`
			Metadata            map[string]string       `json:"metadata,omitempty"`
			FrequencyPenalty    float32                 `json:"frequency_penalty,omitempty"`
			LogitBias           map[string]int          `json:"logit_bias,omitempty"`
			Logprobs            bool                    `json:"logprobs,omitempty"`
			TopLogprobs         *int                    `json:"top_logprobs,omitempty"`
			MaxCompletionTokens int                     `json:"max_completion_tokens,omitempty"`
			N                   int                     `json:"n,omitempty"`
			Modalities          []string                `json:"modalities,omitempty"`
			Prediction          *Prediction             `json:"prediction,omitempty"`
			Audio               *ChatCompletionReqAudio `json:"audio,omitempty"`
			PresencePenalty     int                     `json:"presence_penalty,omitempty"`
			ResponseFormat      *ResponseFormat         `json:"response_format,omitempty"`
			Seed                int                     `json:"seed,omitempty"`
			ServiceTier         string                  `json:"service_tier,omitempty"`
			Stop                []string                `json:"stop,omitempty"`
			Stream              bool                    `json:"stream,omitempty"`
			StreamOptions       *StreamOptions          `json:"stream_options,omitempty"`
			Temperature         *float32                `json:"temperature,omitempty"`
			TopP                *float32                `json:"top_p,omitempty"`
			Tools               []Tool                  `json:"tools,omitempty"`
			ToolChoice          string                  `json:"tool_choice,omitempty"`
			ToolChoiceObj       *ToolChoice             `json:"-"`
			ParallelToolCalls   *bool                   `json:"parallel_tool_calls,omitempty"`
			User                string                  `json:"user,omitempty"`
		}(c)
		return json.Marshal(&req)
	} else {
		req := struct {
			Model               string                  `json:"model"`
			Messages            []Message               `json:"messages"`
			Store               bool                    `json:"store,omitempty"`
			ReasoningEffort     string                  `json:"reasoning_effort,omitempty"`
			Metadata            map[string]string       `json:"metadata,omitempty"`
			FrequencyPenalty    float32                 `json:"frequency_penalty,omitempty"`
			LogitBias           map[string]int          `json:"logit_bias,omitempty"`
			Logprobs            bool                    `json:"logprobs,omitempty"`
			TopLogprobs         *int                    `json:"top_logprobs,omitempty"`
			MaxCompletionTokens int                     `json:"max_completion_tokens,omitempty"`
			N                   int                     `json:"n,omitempty"`
			Modalities          []string                `json:"modalities,omitempty"`
			Prediction          *Prediction             `json:"prediction,omitempty"`
			Audio               *ChatCompletionReqAudio `json:"audio,omitempty"`
			PresencePenalty     int                     `json:"presence_penalty,omitempty"`
			ResponseFormat      *ResponseFormat         `json:"response_format,omitempty"`
			Seed                int                     `json:"seed,omitempty"`
			ServiceTier         string                  `json:"service_tier,omitempty"`
			Stop                []string                `json:"stop,omitempty"`
			Stream              bool                    `json:"stream,omitempty"`
			StreamOptions       *StreamOptions          `json:"stream_options,omitempty"`
			Temperature         *float32                `json:"temperature,omitempty"`
			TopP                *float32                `json:"top_p,omitempty"`
			Tools               []Tool                  `json:"tools,omitempty"`
			ToolChoice          string                  `json:"-"`
			ToolChoiceObj       *ToolChoice             `json:"tool_choice,omitempty"`
			ParallelToolCalls   *bool                   `json:"parallel_tool_calls,omitempty"`
			User                string                  `json:"user,omitempty"`
		}(c)
		return json.Marshal(&req)
	}
}

type ToolChoice struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type ResponseFormat struct {
	Type       string      `json:"type"`
	JsonSchema *JsonSchema `json:"json_schema,omitempty"`
}

type JsonSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
	Strict      bool           `json:"strict,omitempty"`
}

type ChatCompletionReqAudio struct {
	Voice  string `json:"voice"`
	Format string `json:"format"`
}

type Prediction struct {
	Type              string `json:"type"`
	Content           string `json:"content"`
	MultipartContents []MultipartContent
}

func (p *Prediction) UnmarshalJSON(bytes []byte) (err error) {
	{
		var req struct {
			Type              string             `json:"type"`
			Content           string             `json:"content"`
			MultipartContents []MultipartContent `json:"-"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*p = Prediction(req)
			return nil
		}
	}

	{
		var req struct {
			Type              string             `json:"type"`
			Content           string             `json:"-"`
			MultipartContents []MultipartContent `json:"content"`
		}
		if err = json.Unmarshal(bytes, &req); err == nil {
			*p = Prediction(req)
			return nil
		}
	}
	return
}

func (p Prediction) MarshalJSON() ([]byte, error) {
	if (p.Content == "" && len(p.MultipartContents) == 0) || p.Content != "" {
		req := struct {
			Type              string             `json:"type"`
			Content           string             `json:"content"`
			MultipartContents []MultipartContent `json:"-"`
		}(p)
		return json.Marshal(&req)
	} else {
		req := struct {
			Type              string             `json:"type"`
			Content           string             `json:"-"`
			MultipartContents []MultipartContent `json:"content"`
		}(p)
		return json.Marshal(&req)
	}
}

type ChatCompletionResp struct {
	ID                string   `json:"id"`
	Choices           []Choice `json:"choices"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	ServiceTier       string   `json:"service_tier,omitempty"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
	Object            string   `json:"object,omitempty"`
	Usage             Usage    `json:"usage,omitempty"`
	Error             *Error   `json:"error,omitempty"`
}

type Usage struct {
	CompletionTokens        int `json:"completion_tokens"`
	PromptTokens            int `json:"prompt_tokens"`
	TotalTokens             int `json:"total_tokens"`
	CompletionTokensDetails struct {
		AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
		AudioTokens              int `json:"audio_tokens"`
		ReasoningTokens          int `json:"reasoning_tokens"`
		RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	} `json:"completion_tokens_details"`
	PromptTokensDetails struct {
		AudioTokens  int `json:"audio_tokens"`
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
}

type ChoiceLogprobsContent struct {
	Token       string  `json:"token"`
	Logprob     float32 `json:"logprob"`
	Bytes       []int   `json:"bytes"`
	TopLogprobs []struct {
		Token   string  `json:"token"`
		Logprob float32 `json:"logprob"`
		Bytes   []int   `json:"bytes"`
	} `json:"top_logprobs"`
}

type Choice struct {
	FinishReason string   `json:"finish_reason"`
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"`
	Logprobs     *struct {
		Content []ChoiceLogprobsContent `json:"content"`
		refusal []ChoiceLogprobsContent `json:"refusal"`
	} `json:"logprobs,omitempty"`
}

type Message struct {
	Role              string `json:"role"`
	Content           string `json:"content,omitempty"`
	MultipartContents []MultipartContent
	Name              string        `json:"name,omitempty"`
	ToolCalls         []ToolCall    `json:"tool_calls,omitempty"`
	ToolCallID        string        `json:"tool_call_id,omitempty"`
	Refusal           string        `json:"refusal,omitempty"`
	Audio             *MessageAudio `json:"audio,omitempty"`
}

func (m *Message) UnmarshalJSON(bytes []byte) (err error) {
	{
		var req struct {
			Role              string             `json:"role"`
			Content           string             `json:"content,omitempty"`
			MultipartContents []MultipartContent `json:"-"`
			Name              string             `json:"name,omitempty"`
			ToolCalls         []ToolCall         `json:"tool_calls,omitempty"`
			ToolCallID        string             `json:"tool_call_id,omitempty"`
			Refusal           string             `json:"refusal,omitempty"`
			Audio             *MessageAudio      `json:"audio,omitempty"`
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
			MultipartContents []MultipartContent `json:"content,omitempty"`
			Name              string             `json:"name,omitempty"`
			ToolCalls         []ToolCall         `json:"tool_calls,omitempty"`
			ToolCallID        string             `json:"tool_call_id,omitempty"`
			Refusal           string             `json:"refusal,omitempty"`
			Audio             *MessageAudio      `json:"audio,omitempty"`
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
			Content           string             `json:"content,omitempty"`
			MultipartContents []MultipartContent `json:"-"`
			Name              string             `json:"name,omitempty"`
			ToolCalls         []ToolCall         `json:"tool_calls,omitempty"`
			ToolCallID        string             `json:"tool_call_id,omitempty"`
			Refusal           string             `json:"refusal,omitempty"`
			Audio             *MessageAudio      `json:"audio,omitempty"`
		}(m)
		return json.Marshal(&req)
	} else {
		req := struct {
			Role              string             `json:"role"`
			Content           string             `json:"-"`
			MultipartContents []MultipartContent `json:"content,omitempty"`
			Name              string             `json:"name,omitempty"`
			ToolCalls         []ToolCall         `json:"tool_calls,omitempty"`
			ToolCallID        string             `json:"tool_call_id,omitempty"`
			Refusal           string             `json:"refusal,omitempty"`
			Audio             *MessageAudio      `json:"audio,omitempty"`
		}(m)
		return json.Marshal(&req)
	}
}

type MessageAudio struct {
	ID string `json:"id"`
}

type MultipartContent struct {
	Type       string     `json:"type"`
	Text       string     `json:"text,omitempty"`
	ImageUrl   ImageUrl   `json:"image_url,omitempty"`
	InputAudio InputAudio `json:"input_audio,omitempty"`
}

type ImageUrl struct {
	Url    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type Tool struct {
	Type     client.ToolType `json:"type"`
	Function Function        `json:"function"`
}

type Function struct {
	Name        string     `json:"name"`
	Arguments   string     `json:"arguments,omitempty"`
	Description string     `json:"description,omitempty"`
	Parameters  Parameters `json:"parameters,omitempty"`
	Strict      bool       `json:"strict,omitempty"`
}

type Parameters struct {
	Type                 string                     `json:"type"`
	Properties           map[string]client.Property `json:"properties"`
	Required             []string                   `json:"required"`
	AdditionalProperties bool                       `json:"additionalProperties"`
}

type ToolCall struct {
	ID       string          `json:"id,omitempty"`
	Type     client.ToolType `json:"type,omitempty"`
	Function Function        `json:"function,omitempty"`
}

type CreateBatchReq struct {
	InputFileID      string         `json:"input_file_id"`
	Endpoint         string         `json:"endpoint"`
	CompletionWindow string         `json:"completion_window"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// BatchResponse 表示 OpenAI 批处理任务的返回值
type BatchResponse struct {
	ID               string        `json:"id"`                       // 批处理任务的唯一标识符
	Object           string        `json:"object"`                   // 对象类型，通常为 "batch"
	Endpoint         string        `json:"endpoint"`                 // 请求的 API 端点
	Errors           *ErrorDetails `json:"errors,omitempty"`         // 错误详情（如果有）
	InputFileID      string        `json:"input_file_id"`            // 输入文件的 ID
	CompletionWindow string        `json:"completion_window"`        // 任务完成的时间窗口
	Status           BatchStatus   `json:"status"`                   // 任务状态（如 "completed", "failed"）
	OutputFileID     *string       `json:"output_file_id,omitempty"` // 输出文件的 ID（任务完成后）
	ErrorFileID      *string       `json:"error_file_id,omitempty"`  // 错误文件的 ID（如果有错误）
	CreatedAt        int64         `json:"created_at"`               // 任务创建时间（Unix 时间戳）
	CompletedAt      *int64        `json:"completed_at,omitempty"`   // 任务完成时间（Unix 时间戳）
}

type BatchStatus string

const (
	Validating BatchStatus = "validating"  //the input file is being validated before the batch can begin
	Failed     BatchStatus = "failed"      //the input file has failed the validation process
	InProgress BatchStatus = "in_progress" //the input file was successfully validated and the batch is currently being run
	Finalizing BatchStatus = "finalizing"  //the batch has completed and the results are being prepared
	Completed  BatchStatus = "completed"   //the batch has been completed and the results are ready
	Expired    BatchStatus = "expired"     //batch was not able to be completed within the 24-hour time window
	Cancelling BatchStatus = "cancelling"  //the batch is being cancelled (may take up to 10 minutes)
	Cancelled  BatchStatus = "cancelled"   //the batch was cancelled
)

// ErrorDetails 表示批处理任务中的错误详情
type ErrorDetails struct {
	Object string      `json:"object"` // 对象类型，通常为 "list"
	Data   []ErrorData `json:"data"`   // 错误数据列表
}

// ErrorData 表示单个错误的详细信息
type ErrorData struct {
	Code    string `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误消息
	Param   string `json:"param"`   // 错误参数（如果有）
}

type BatchReqItem struct {
	CustomId string           `json:"custom_id"`
	Method   string           `json:"method"`
	Url      string           `json:"url"`
	Body     BatchReqItemBody `json:"body"`
}

type BatchReqItemBody struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type UploadResp struct {
	ID        string `json:"id"`
	Bytes     int    `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	Filename  string `json:"filename"`
	Object    string `json:"object"`
	Purpose   string `json:"purpose"`
}
