package zhipu

import (
	"github.com/goslacker/slacker/ai"
	"github.com/goslacker/slacker/tool"
)

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Enum        []any
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Function struct {
	Name        string `json:"name"`        //函数名称，只能包含a-z，A-Z，0-9，下划线和中横线。最大长度限制为64
	Description string `json:"description"` //用于描述函数功能。模型会根据这段描述决定函数调用方式。

	//parameter 字段需要传入一个 Json Schema 对象，以准确地定义函数所接受的参数。样例：
	//  "parameters": {
	//    "type": "object",
	//    "properties": {
	//      "location": {
	//        "type": "string",
	//        "description": "城市，如：北京",
	//      },
	//      "unit": {
	//        "type": "string",
	//        "enum": ["celsius", "fahrenheit"]
	//      },
	//    },
	//    "required": ["location"],
	//  }
	Parameters Parameters `json:"parameters"`
}

type Retrieval struct {
	KnowledgeID string `json:"knowledge_id"` //当涉及到知识库ID时，请前往开放平台的知识库模块进行创建或获取。
	//请求模型时的知识库模板，默认模板：
	//  从文档
	//  """
	//  {{ knowledge}}
	//  """
	//  中找问题
	//  """
	//  {{question}}
	//  """
	//  的答案，找到答案就仅使用文档语句回答问题，找不到答案就用自身知识回答并且告诉用户该信息不是来自文档。
	//
	//  不要复述问题，直接开始回答
	//
	//
	//注意：用户自定义模板时，知识库内容占位符和用户侧问题占位符必是{{ knowledge}} 和{{question}}，其他模板内容用户可根据实际场景定义
	PromptTemplate string `json:"prompt_template,omitempty"`
}

type WebSearch struct {
	Enable       *bool  `json:"enable,omitempty"`
	SearchQuery  string `json:"search_query,omitempty"`
	SearchResult *bool  `json:"search_result,omitempty"`
}

type Tool struct {
	Type      string     `json:"type"`                 // 工具类型,目前支持function、retrieval、web_search
	Function  *Function  `json:"function,omitempty"`   //仅当工具类型为function时补充
	Retrieval *Retrieval `json:"retrieval,omitempty"`  //仅当工具类型为retrieval时补充
	WebSearch *WebSearch `json:"web_search,omitempty"` //仅当工具类型为web_search时补充，如果tools中存在类型retrieval，此时web_search不生效。
}

type ChatCompletionReq struct {
	Model     string    `json:"model" validate:"required"`    //所要调用的模型编码
	Messages  []Message `json:"messages" validate:"required"` //调用语言模型时，将当前对话信息列表作为提示输入给模型， 按照 {"role": "user", "content": "你好"} 的json 数组形式进行传参； 可能的消息类型包括 System message、User message、Assistant message 和 Tool message。
	RequestID *string   `json:"request_id,omitempty"`         //由用户端传参，需保证唯一性；用于区分每次请求的唯一标识，用户端不传时平台会默认生成。
	DoSimple  *bool     `json:"do_sample,omitempty"`          //为 true 时启用采样策略，do_sample 为 false 时采样策略 temperature、top_p 将不生效。默认值为 true。

	//使用同步调用时，此参数应当设置为 fasle 或者省略。表示模型生成完所有内容后一次性返回所有内容。默认值为 false。
	//
	//如果设置为 true，模型将通过标准 Event Stream ，逐块返回模型生成内容。Event Stream 结束时会返回一条data: [DONE]消息。
	//
	//注意：在模型流式输出生成内容的过程中，我们会分批对模型生成内容进行检测，当检测到违法及不良信息时，API会返回错误码（1301）。开发者识别到错误码（1301），应及时采取（清屏、重启对话）等措施删除生成内容，并确保不将含有违法及不良信息的内容传递给模型继续生成，避免其造成负面影响。
	Stream *bool `json:"stream,omitempty"`

	//采样温度，控制输出的随机性，必须为正数
	//
	//取值范围是：(0.0, 1.0)，不能等于 0，默认值为 0.95，值越大，会使输出更随机，更具创造性；值越小，输出会更加稳定或确定
	//
	//建议您根据应用场景调整 top_p 或 temperature 参数，但不要同时调整两个参数
	Temperature float32 `json:"temperature,omitempty"`

	//用温度取样的另一种方法，称为核取样
	//
	//取值范围是：(0.0, 1.0) 开区间，不能等于 0 或 1，默认值为 0.7
	//
	//模型考虑具有 top_p 概率质量 tokens 的结果
	//
	//例如：0.1 意味着模型解码器只考虑从前 10% 的概率的候选集中取 tokens
	//
	//建议您根据应用场景调整 top_p 或 temperature 参数，但不要同时调整两个参数
	TopP      float32  `json:"top_p,omitempty"`
	MaxTokens int      `json:"max_tokens,omitempty"` //模型输出最大 tokens，最大输出为4095，默认值为1024
	Stop      []string `json:"stop,omitempty"`       //模型在遇到stop所制定的字符时将停止生成，目前仅支持单个停止词，格式为["stop_word1"]

	Tools []Tool `json:"tools,omitempty"` //可供模型调用的工具。默认开启web_search ，调用成功后作为参考信息提供给模型。注意：返回结果作为输入也会进行计量计费，每次调用大约会增加1000 tokens的消耗。

	ToolChoice string `json:"tool_choice,omitempty"` //用于控制模型是如何选择要调用的函数，仅当工具类型为function时补充。默认为auto，当前仅支持auto

	//终端用户的唯一ID，协助平台对终端用户的违规行为、生成违法及不良信息或其他滥用行为进行干预。ID长度要求：最少6个字符，最多128个字符。
	//see https://open.bigmodel.cn/dev/howuse/securityaudit
	UserID string `json:"user_id,omitempty"`
}

type FunctionCallInfo struct {
	Name      string //函数名称
	Arguments string //模型生成的调用函数的参数列表，json 格式。请注意，模型可能会生成无效的JSON，也可能会虚构一些不在您的函数规范中的参数。在调用函数之前，请在代码中验证这些参数是否有效。
}

type Message struct {
	Role      string `json:"role"`              //消息的角色信息，此时应为system user assistant tool
	Content   string `json:"content,omitempty"` //消息内容
	ToolCalls []struct {
		ID       string            `json:"id"`                 //工具id
		Type     string            `json:"type"`               //工具类型,支持web_search、retrieval、function
		Function *FunctionCallInfo `json:"function,omitempty"` //type为"function"时不为空
	} `json:"tool_calls,omitempty"` //模型产生的工具调用消息
	ToolCallID string `json:"tool_call_id,omitempty"` //tool的调用记录
}

func MessageFromStandard(m *ai.Message) (message *Message, err error) {
	err = tool.SimpleMap(&message, m)
	return
}

type Choice struct {
	Index int `json:"index"` //结果下标
	//模型推理终止的原因。
	//  stop代表推理自然结束或触发停止词。
	//  tool_calls 代表模型命中函数。
	//  length代表到达 tokens 长度上限。
	//  sensitive 代表模型推理内容被安全审核接口拦截。请注意，针对此类内容，请用户自行判断并决定是否撤回已公开的内容。
	//  network_error 代表模型推理异常。
	FinishReason string   `json:"finish_reason"`
	Delta        *Message `json:"delta"`
	Message      *Message `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     //用户输入的 tokens 数量
	CompletionTokens int `json:"completion_tokens"` //模型输出的 tokens 数量
	TotalTokens      int `json:"total_tokens"`      //总 tokens 数量
}

type WebSearchResp struct {
	Icon    string `json:"icon"`    //来源网站的icon
	Title   string `json:"title"`   //搜索结果的标题
	Link    string `json:"link"`    //搜索结果的网页链接
	Media   string `json:"media"`   //搜索结果网页来源的名称
	Content string `json:"content"` //从搜索结果网页中引用的文本内容
}

type ChatCompletionResp struct {
	ID        string        `json:"id"`
	Created   int64         `json:"created"`
	Model     string        `json:"model"`
	Choices   []Choice      `json:"choices"`
	Usage     Usage         `json:"usage"`
	WebSearch WebSearchResp `json:"web_search"`
}
