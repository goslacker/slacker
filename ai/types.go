package ai

type FunctionCallInfo struct {
	Name      string //函数名称
	Arguments string //模型生成的调用函数的参数列表，json 格式。请注意，模型可能会生成无效的JSON，也可能会虚构一些不在您的函数规范中的参数。在调用函数之前，请在代码中验证这些参数是否有效。
}

type Message struct {
	Role      string `json:"role"`    //消息的角色信息，system user assistant tool
	Content   string `json:"content"` //消息内容
	ToolCalls []struct {
		ID       string            //id
		Type     string            //类型
		Function *FunctionCallInfo //function描述 type为"function"时不为空
	} `json:"-"` //模型产生的工具调用消息
	ToolCallID string `json:"-"` //tool的调用记录
}
