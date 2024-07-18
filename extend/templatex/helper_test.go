package templatex

import (
	"github.com/stretchr/testify/require"
	"testing"
	"text/template"
)

func TestNormal(t *testing.T) {
	products := []string{
		"便当蛋糕：高颜值，口味丰富，款式多样，方便保存。",
		"蛋糕卷：下午茶的好选择，可以切块，口味稳定，做法简单。",
		"千层蛋糕：网红甜品，口味多样，大众认可，可以切块。",
		"杯子蛋糕：适合打卡，口味丰富，高颜值，方便保存。",
		"菠萝包：经典产品，长期售卖，口味稳定，方便保存。",
		"盒子蛋糕：饱腹感强，方便保存，口味多样，做法简单。",
		"燃烧土豆：大众认可，方便保存，饱腹感强，经典产品。",
		"法式可可豆：适合售卖，甜而不腻，高颜值，方便保存。",
		"雪媚娘：经典产品，长期售卖，口味稳定，容易保存。",
		"盐可颂：饱腹感强，容易保存，适合售卖，做法简单。",
		"曲奇：下午茶好选择，可做礼品，长期售卖，保质期长。",
		"蛋黄酥：饱腹感强，甜而不腻，可做礼品，容易保存。",
		"芝士咖喱鸡：经典产品，口味浓郁，饱腹感强，容易保存。",
		"布朗尼：饱腹感强，容易保存，适合售卖，做法简单。",
		"柠檬蛋糕：下午茶好选择，可做礼品，长期售卖，保质期长。",
		"麻薯：Q弹软糯，容易保存，携带便利，私房必备。",
	}
	temp := NewTextTemplate(`你是一个烘培店的销售客服，你负责解答客户的问题以及推销自己店铺的产品。
当客户不知道自己该买什么时，你应该根据客户的喜好向顾客推荐我们的产品。
如果你不知道客户的喜好，**不要假设或猜测**，请要求客户提供必要信息。
当你需要向客户推荐我们的产品时，你需要从我们已有的产品中选择你的推荐目标。
你应该使用中文来回复客户。
产品信息如下:
{{.ProductList}}`)
	temp.AddSub("ProductList", NewTextTemplate(`我们有这些产品：
{{ range $index, $item := .Products }}{{ add1 $index }}. {{ $item }}
{{ end }}`))
	result, err := temp.RenderWithFuncMap(map[string]any{
		"Products": products,
	}, template.FuncMap{
		"add1": func(i int) int {
			return i + 1
		},
	})
	require.NoError(t, err)
	print(result)
}
