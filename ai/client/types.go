package client

type AIClient interface {
	ChatCompletion(req *ChatCompletionReq) (resp *ChatCompletionResp, err error)
}

type ChatCompletionReq struct {
	Model             string            `json:"model"`
	Messages          []Message         `json:"messages"`
	FrequencyPenalty  float32           `json:"frequencyPenalty,omitempty"`
	LogitBias         map[string]int    `json:"logitBias,omitempty"`
	Logprobs          bool              `json:"logprobs,omitempty"`
	TopLogprobs       int               `json:"topLogprobs,omitempty"`
	MaxTokens         int               `json:"maxTokens,omitempty"`
	N                 int               `json:"n,omitempty"`
	PresencePenalty   int               `json:"presencePenalty,omitempty"`
	ResponseFormat    map[string]string `json:"responseFormat,omitempty"`
	Seed              int               `json:"seed,omitempty"`
	ServiceTier       string            `json:"serviceTier,omitempty"`
	Stop              []string          `json:"stop,omitempty"`
	Stream            bool              `json:"stream,omitempty"`
	StreamOptions     map[string]any    `json:"stream_options,omitempty"`
	Temperature       *float32          `json:"temperature,omitempty"`
	TopP              *float32          `json:"top_p,omitempty"`
	Tools             []Tool            `json:"tools,omitempty"`
	ToolChoice        any               `json:"tool_choice,omitempty"` //string || []ToolChoice
	ParallelToolCalls *bool             `json:"disable_parallel_tool_calls,omitempty"`
	User              string            `json:"user,omitempty"`
}

type ChatCompletionResp struct {
	ID                string          `json:"id,omitempty"`
	Choices           []Choice        `json:"choices"`
	Created           int64           `json:"created,omitempty"`
	Model             string          `json:"model"`
	ServiceTier       string          `json:"service_tier,omitempty"`
	SystemFingerprint string          `json:"system_fingerprint,omitempty"`
	Object            string          `json:"object,omitempty"`
	Usage             Usage           `json:"usage,omitempty"`
	WebSearch         []WebSearchResp `json:"web_search,omitempty"`
}

type WebSearchResp struct {
	Icon    string `json:"icon"`    //来源网站的icon
	Title   string `json:"title"`   //搜索结果的标题
	Link    string `json:"link"`    //搜索结果的网页链接
	Media   string `json:"media"`   //搜索结果网页来源的名称
	Content string `json:"content"` //从搜索结果网页中引用的文本内容
}

type Usage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	FinishReason string  `json:"finish_reason,omitempty"`
	Index        int     `json:"index,omitempty"`
	Message      Message `json:"message,omitempty"`
	Delta        Message `json:"delta,omitempty"`
	Logprobs     *struct {
		Content []struct {
			Token       string  `json:"token"`
			Logprob     float32 `json:"logprob"`
			Bytes       []int   `json:"bytes"`
			TopLogprobs []struct {
				Token   string  `json:"token"`
				Logprob float32 `json:"logprob"`
				Bytes   []int   `json:"bytes"`
			} `json:"top_logprobs"`
		} `json:"content"`
	} `json:"logprobs,omitempty"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    any        `json:"content,omitempty"` // string || []Content
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolType string

const (
	ToolTypeFunction  ToolType = "function"
	ToolTypeRetrieval ToolType = "retrieval"
	ToolTypeWebSearch ToolType = "web_search"
)

type Tool struct {
	Type      ToolType   `json:"type"`
	Function  *Function  `json:"function,omitempty"`
	Retrieval *Retrieval `json:"retrieval,omitempty"`
	WebSearch *WebSearch `json:"web_search,omitempty"`
}

type WebSearch struct {
	Enable       *bool  `json:"enable,omitempty"`
	SearchQuery  string `json:"search_query"`
	SearchResult bool   `json:"search_result"`
}

type Retrieval struct {
	KnowledgeID    string `json:"knowledge_id"`
	PromptTemplate string `json:"prompt_template"`
}

type Function struct {
	Description string      `json:"description,omitempty"`
	Name        string      `json:"name,omitempty"`
	Parameters  *Parameters `json:"parameters,omitempty"`
	Arguments   string      `json:"arguments,omitempty"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeImageUrl ContentType = "image_url"
)

type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageUrl string `json:"image_url,omitempty"`
}

type ToolCall struct {
	ID       string   `json:"id,omitempty"`
	Type     ToolType `json:"type,omitempty"`
	Function Function `json:"function,omitempty"`
}
