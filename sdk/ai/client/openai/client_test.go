package openai

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/goslacker/slacker/sdk/ai/client"

	"github.com/stretchr/testify/require"
)

const key = ""
const apiKey = ""

func genPwd(saltBytes []byte) string {
	hmacHash := hmac.New(sha1.New, []byte(key))
	hash := hmacHash.Sum(saltBytes)

	offset := hash[len(hash)-1] & 0x0F
	truncatedHash := hash[offset : offset+4]
	code := binary.BigEndian.Uint32(truncatedHash) & 0x7FFFFFFF
	return strconv.FormatUint(uint64(code), 10)
}

func TestClient(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	saltBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(saltBytes, uint64(time.Now().Unix()))
	//code := genPwd(saltBytes)
	c := NewClient(apiKey, client.WithBaseUrl("http:///v1"), client.WithHttpHeader(http.Header{
		"X-Forwarded-Host": []string{"https://api.openai.com"},
	}))
	resp, err := c.ChatCompletion(&client.ChatCompletionReq{
		Model:          "gpt-4o-2024-11-20",
		ResponseFormat: map[string]string{"type": "json_object"},
		Messages: []client.Message{
			{
				Role:    "system",
				Content: "您应该始终遵循指令并输出一个有效的JSON对象。请根据指令使用指定的JSON对象结构。如果不确定，请默认使用 {\"answer\": \"$your_answer\"}。确保始终不以 \"```\" 结束代码块。",
			},
			{
				Role:    "user",
				Content: "你好",
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	b, _ := json.Marshal(resp)
	println(string(b))
}

var productList = `[
  {
    "name": "粉色祝寿蛋糕",
    "size": "6寸",
    "price": 328,
    "recipient_tag": "老人",
    "flavor_tag": "芒果",
    "healthy_tag": "无糖,低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/369dvesutqbtl13",
    "product_image_url": "https://img01.yzcdn.cn/upload_files/2024/07/02/Ftv4SMM_zAd3pUSlaNspdvakELJy.jpg!large.webp"
  },
  {
    "name": "粉色祝寿蛋糕",
    "size": "8寸",
    "price": 398,
    "recipient_tag": "送老人",
    "flavor_tag": "芒果",
    "healthy_tag": "无糖,低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/369dvesutqbtl13",
    "product_image_url": "https://img01.yzcdn.cn/upload_files/2024/07/02/Ftv4SMM_zAd3pUSlaNspdvakELJy.jpg!large.webp"
  },
  {
    "name": "兔兔周岁蛋糕",
    "size": "6寸",
    "price": 258,
    "recipient_tag": "送小女孩",
    "flavor_tag": "蓝莓，巧克力",
    "healthy_tag": "低糖，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/26wwjdp2akq89va",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/02/Fg39OBYX-77zwWWURHzbCcTOYDHK.jpg!large.webp"
  },
  {
    "name": "兔兔周岁蛋糕",
    "size": "8寸",
    "price": 328,
    "recipient_tag": "送小女孩",
    "flavor_tag": "蓝莓，巧克力",
    "healthy_tag": "低糖，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/26wwjdp2akq89va",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/02/Fg39OBYX-77zwWWURHzbCcTOYDHK.jpg!large.webp"
  },
  {
    "name": "JOJO卡通汽车蛋糕",
    "size": "6寸加高",
    "price": 268,
    "recipient_tag": "送小男孩",
    "flavor_tag": "芋泥",
    "healthy_tag": "低糖，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3njf7bp8v2kmhv2",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/26/Fh76ZZIdIr7ya3hkUF31rxQZ6zKk.jpg!large.webp"
  },
  {
    "name": "JOJO卡通汽车蛋糕",
    "size": "8寸加高",
    "price": 338,
    "recipient_tag": "送小男孩",
    "flavor_tag": "芋泥",
    "healthy_tag": "低糖，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3njf7bp8v2kmhv2",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/26/Fh76ZZIdIr7ya3hkUF31rxQZ6zKk.jpg!large.webp"
  },
  {
    "name": "熊熊夫妻纪念日蛋糕",
    "size": "6寸",
    "price": 188,
    "recipient_tag": "送爱人",
    "flavor_tag": "奥利奥",
    "healthy_tag": "纯动物奶油",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/365nsxt2mijyxkf",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/26/Fv0yNNCDFhiChF4kZgGeizFUEhAi.jpg!large.webp"
  },
  {
    "name": "复古蝴蝶结蛋糕",
    "size": "6寸",
    "price": 198,
    "recipient_tag": "送爱人",
    "flavor_tag": "蓝莓",
    "healthy_tag": "纯动物奶油",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/1ybpsbz25yd2xen",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/26/Fm9TtQKCzOp6aqBiqckKmeEVSf2m.jpg!large.webp"
  },
  {
    "name": "复古蝴蝶结蛋糕",
    "size": "8寸",
    "price": 268,
    "recipient_tag": "送爱人",
    "flavor_tag": "蓝莓",
    "healthy_tag": "纯动物奶油",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/1ybpsbz25yd2xen",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/26/Fm9TtQKCzOp6aqBiqckKmeEVSf2m.jpg!large.webp"
  },
  {
    "name": "蝴蝶蛋糕",
    "size": "6寸",
    "price": 298,
    "recipient_tag": "送女神",
    "flavor_tag": "巧克力，奥利奥，芋泥",
    "healthy_tag": "纯动物奶油，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2oecfzt4uakaxhb",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/24/FkM1Jl3c6TCMPdIdiR-A9g2aPLKw.jpg!large.webp"
  },
  {
    "name": "蝴蝶蛋糕",
    "size": "8寸",
    "price": 368,
    "recipient_tag": "送女神",
    "flavor_tag": "巧克力，奥利奥，芋泥",
    "healthy_tag": "纯动物奶油，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2oecfzt4uakaxhb",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/24/FkM1Jl3c6TCMPdIdiR-A9g2aPLKw.jpg!large.webp"
  },
  {
    "name": "荔枝梅梅",
    "size": "5寸",
    "price": 178,
    "recipient_tag": "送女神",
    "flavor_tag": "荔枝，树莓",
    "healthy_tag": "纯动物奶油，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/26ufi3j29w20psr",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/24/FiHXnaY_-eOjvL9LEBRzQS6WviZL.jpg!large.webp"
  },
  {
    "name": "荔枝梅梅",
    "size": "6寸",
    "price": 228,
    "recipient_tag": "送女神",
    "flavor_tag": "荔枝，树莓",
    "healthy_tag": "纯动物奶油，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/26ufi3j29w20psr",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/24/FiHXnaY_-eOjvL9LEBRzQS6WviZL.jpg!large.webp"
  },
  {
    "name": "荔枝梅梅",
    "size": "8寸",
    "price": 298,
    "recipient_tag": "送女神",
    "flavor_tag": "荔枝，树莓",
    "healthy_tag": "纯动物奶油，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/26ufi3j29w20psr",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/24/FiHXnaY_-eOjvL9LEBRzQS6WviZL.jpg!large.webp"
  },
  {
    "name": "乌梅子酱蝴蝶兰",
    "size": "6寸",
    "price": 218,
    "recipient_tag": "送女神",
    "flavor_tag": "乌梅酱，蓝莓",
    "healthy_tag": "纯动物奶油，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/276qroxzww655gc",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/20/FsOk-KpUfAuzaWV9GC0p133s2M6g.jpg!large.webp"
  },
  {
    "name": "乌梅子酱蝴蝶兰",
    "size": "8寸",
    "price": 288,
    "recipient_tag": "送女神",
    "flavor_tag": "乌梅酱，蓝莓",
    "healthy_tag": "纯动物奶油，木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/276qroxzww655gc",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/20/FsOk-KpUfAuzaWV9GC0p133s2M6g.jpg!large.webp"
  },
  {
    "name": "小王子蛋糕",
    "size": "6寸",
    "price": 238,
    "recipient_tag": "送小男孩",
    "flavor_tag": "芋泥，巧克力",
    "healthy_tag": "低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3nlubo55ocbuxmb",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/20/Fi6y0S9bYqm9o_EeiKG6VnwjXobg.jpg!large.webp"
  },
  {
    "name": "小王子蛋糕",
    "size": "8寸",
    "price": 298,
    "recipient_tag": "送小男孩",
    "flavor_tag": "芋泥，巧克力",
    "healthy_tag": "低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3nlubo55ocbuxmb",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/20/Fi6y0S9bYqm9o_EeiKG6VnwjXobg.jpg!large.webp"
  },
  {
    "name": "双面蛋糕",
    "size": "6寸",
    "price": 218,
    "recipient_tag": "送男神",
    "flavor_tag": "巧克力",
    "healthy_tag": "纯动物奶油",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2oo8exntku6a1rv",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/19/FutCgP4JZxpk6UhC-HGT0l36Tax-.jpg!large.webp"
  },
  {
    "name": "双面蛋糕",
    "size": "8寸",
    "price": 288,
    "recipient_tag": "送男神",
    "flavor_tag": "巧克力",
    "healthy_tag": "纯动物奶油",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2oo8exntku6a1rv",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/19/FutCgP4JZxpk6UhC-HGT0l36Tax-.jpg!large.webp"
  },
  {
    "name": "小脑虎蛋糕",
    "size": "6寸",
    "price": 298,
    "recipient_tag": "送小男孩",
    "flavor_tag": "芒果",
    "healthy_tag": "无糖，低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/271u97dgtljjtyg",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/18/FtkLZJzBOR3K4rPPTjSpl09Vq6wm.jpg!large.webp"
  },
  {
    "name": "小脑虎蛋糕",
    "size": "8寸",
    "price": 368,
    "recipient_tag": "送小男孩",
    "flavor_tag": "芒果",
    "healthy_tag": "无糖，低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/271u97dgtljjtyg",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/18/FtkLZJzBOR3K4rPPTjSpl09Vq6wm.jpg!large.webp"
  },
  {
    "name": "水中月",
    "size": "4寸",
    "price": 128,
    "recipient_tag": "送爱人",
    "flavor_tag": "奥利奥",
    "healthy_tag": "木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2oee6jrkhjezdqy",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/29/FsJvoGS0josQTt3iH7iYERD70UvA.jpg!large.webp"
  },
  {
    "name": "水中月",
    "size": "5寸",
    "price": 178,
    "recipient_tag": "送爱人",
    "flavor_tag": "奥利奥",
    "healthy_tag": "木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2oee6jrkhjezdqy",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/29/FsJvoGS0josQTt3iH7iYERD70UvA.jpg!large.webp"
  },
  {
    "name": "水中月",
    "size": "6寸",
    "price": 228,
    "recipient_tag": "送爱人",
    "flavor_tag": "奥利奥",
    "healthy_tag": "木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2oee6jrkhjezdqy",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/29/FsJvoGS0josQTt3iH7iYERD70UvA.jpg!large.webp"
  },
  {
    "name": "粉色鲜花蛋糕",
    "size": "6寸",
    "price": 198,
    "recipient_tag": "送爱人",
    "flavor_tag": "阳光玫瑰，芋泥",
    "healthy_tag": "木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/36eaq90dccqzd2p",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/20/Fk1_0l-06-8FntBmib42PsQHEZNK.jpg!large.webp"
  },
  {
    "name": "粉色鲜花蛋糕",
    "size": "8寸",
    "price": 268,
    "recipient_tag": "送爱人",
    "flavor_tag": "阳光玫瑰，芋泥",
    "healthy_tag": "木糖醇",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/36eaq90dccqzd2p",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/20/Fk1_0l-06-8FntBmib42PsQHEZNK.jpg!large.webp"
  },
  {
    "name": "妈妈蛋糕",
    "size": "6寸",
    "price": 188,
    "recipient_tag": "送妈妈",
    "flavor_tag": "芒果",
    "healthy_tag": "无糖，低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/1y4d3zgaoid61we",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/09/Fk8u0cG4xoYh3Lr92AJ3wmTE3dAj.jpg"
  },
  {
    "name": "妈妈蛋糕",
    "size": "8寸",
    "price": 258,
    "recipient_tag": "送妈妈",
    "flavor_tag": "芒果",
    "healthy_tag": "无糖，低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/1y4d3zgaoid61we",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/09/Fk8u0cG4xoYh3Lr92AJ3wmTE3dAj.jpg"
  },
  {
    "name": "女士鲜花蛋糕",
    "size": "6寸",
    "price": 228,
    "recipient_tag": "送妈妈",
    "flavor_tag": "芒果",
    "healthy_tag": "无糖，低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3nknjkkt9tg21ib",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/11/Fh_FYasDWmyVW9-7UkyQq63QZDEo.jpg"
  },
  {
    "name": "女士鲜花蛋糕",
    "size": "8寸",
    "price": 298,
    "recipient_tag": "送妈妈",
    "flavor_tag": "芒果",
    "healthy_tag": "无糖，低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3nknjkkt9tg21ib",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/11/Fh_FYasDWmyVW9-7UkyQq63QZDEo.jpg"
  },
  {
    "name": "宇航员蛋糕",
    "size": "6寸",
    "price": 198,
    "recipient_tag": "送男神",
    "flavor_tag": "巧克力，奥利奥，芋泥",
    "healthy_tag": "木糖醇，低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3nn485u8523ndi1",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/11/FqsuLc5tw7MyVXW1SXXIl_HtFu4b.jpg!large.webp"
  },
  {
    "name": "宇航员蛋糕",
    "size": "8寸",
    "price": 268,
    "recipient_tag": "送男神",
    "flavor_tag": "巧克力，奥利奥，芋泥",
    "healthy_tag": "木糖醇，低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3nn485u8523ndi1",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/11/FqsuLc5tw7MyVXW1SXXIl_HtFu4b.jpg!large.webp"
  },
  {
    "name": "彩虹蛋糕",
    "size": "5寸",
    "price": 158,
    "recipient_tag": "送爱人",
    "flavor_tag": "芋泥",
    "healthy_tag": "低糖，无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3nfo9jvnpd909am",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/11/For-k0-vhBvpqkt3PGAMXVyPqbXc.jpg"
  },
  {
    "name": "彩虹蛋糕",
    "size": "6寸",
    "price": 208,
    "recipient_tag": "送爱人",
    "flavor_tag": "芋泥",
    "healthy_tag": "低糖，无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3nfo9jvnpd909am",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/11/For-k0-vhBvpqkt3PGAMXVyPqbXc.jpg"
  },
  {
    "name": "爸爸蛋糕",
    "size": "6寸",
    "price": 198,
    "recipient_tag": "送爸爸",
    "flavor_tag": "蓝莓，草莓",
    "healthy_tag": "低糖，无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2oja628hdsjo9c1",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/09/FvMhwg5ipdpEcNnrqHD1PPGcMiab.jpg!large.webp"
  },
  {
    "name": "爸爸蛋糕",
    "size": "8寸",
    "price": 268,
    "recipient_tag": "送爸爸",
    "flavor_tag": "蓝莓，草莓",
    "healthy_tag": "低糖，无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2oja628hdsjo9c1",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/07/09/FvMhwg5ipdpEcNnrqHD1PPGcMiab.jpg!large.webp"
  },
  {
    "name": "水果男款",
    "size": "6寸",
    "price": 188,
    "recipient_tag": "送爸爸",
    "flavor_tag": "草莓",
    "healthy_tag": "纯动物奶油，木糖醇，低糖，无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3635qet95l4kpmh",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/05/24/Fga4-_JUIPGZtXW7xc0j06sAQPJ8.png!large.webp"
  },
  {
    "name": "水果男款",
    "size": "8寸",
    "price": 258,
    "recipient_tag": "送爸爸",
    "flavor_tag": "草莓",
    "healthy_tag": "纯动物奶油，木糖醇，低糖，无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3635qet95l4kpmh",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/05/24/Fga4-_JUIPGZtXW7xc0j06sAQPJ8.png!large.webp"
  },
  {
    "name": "八方来财款",
    "size": "6寸",
    "price": 188,
    "recipient_tag": "送老人",
    "flavor_tag": "芒果，草莓",
    "healthy_tag": "木糖醇，低糖，无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3en5im1lik3cpm0",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/05/24/Fn_dLHEWnmf6KUKER0stiFEPgW_N.png!large.webp"
  },
  {
    "name": "八方来财款",
    "size": "8寸",
    "price": 258,
    "recipient_tag": "送老人",
    "flavor_tag": "芒果，草莓",
    "healthy_tag": "木糖醇，低糖，无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/3en5im1lik3cpm0",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/05/24/Fn_dLHEWnmf6KUKER0stiFEPgW_N.png!large.webp"
  },
  {
    "name": "福袋寿款蛋糕",
    "size": "6寸",
    "price": 218,
    "recipient_tag": "送老人",
    "flavor_tag": "芒果，草莓",
    "healthy_tag": "低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2xj91yxfq6nq1ep",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/05/Fg_f9s0mvIWz8offE7WQjbImAoLQ.jpg!large.webp"
  },
  {
    "name": "福袋寿款蛋糕",
    "size": "8寸",
    "price": 288,
    "recipient_tag": "送老人",
    "flavor_tag": "芒果，草莓",
    "healthy_tag": "低糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2xj91yxfq6nq1ep",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/06/05/Fg_f9s0mvIWz8offE7WQjbImAoLQ.jpg!large.webp"
  },
  {
    "name": "祝寿款蛋糕8",
    "size": "六寸",
    "price": 188,
    "recipient_tag": "送老人",
    "flavor_tag": "芒果，草莓",
    "healthy_tag": "无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2g31be39dbycpe3",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/05/28/Ft9yljJ_TFRltvRDEKYD-NGf2wj_.png!large.webp"
  },
  {
    "name": "祝寿款蛋糕8",
    "size": "八寸",
    "price": 258,
    "recipient_tag": "送老人",
    "flavor_tag": "芒果，草莓",
    "healthy_tag": "无糖",
    "product_detail_url": "https://shop99796961.m.youzan.com/wscgoods/detail/2g31be39dbycpe3",
    "product_image_url":"https://img01.yzcdn.cn/upload_files/2024/05/28/Ft9yljJ_TFRltvRDEKYD-NGf2wj_.png!large.webp"
  },
]`

func TestClient_ChatCompletion(t *testing.T) {
	t.Skip()
	slog.SetLogLoggerLevel(slog.LevelDebug)
	c := NewClient("")
	resp, err := c.ChatCompletion(&client.ChatCompletionReq{
		Model: "glm-4-0520",
		Messages: []client.Message{
			{
				Role: "system",
				Content: `## 概要
你是小江烘培店的客服伙伴小张，现在一个潜在客户找到了你。你将用贴心有温度的服务态度，通过对话来发掘客户的需求，匹配合适的方案并且将推荐理由基于客户的需求情况针对性的阐述理由，从而促使客户下单购买。在解决了客户当前的问题后，你擅长将对话的焦点从无关的话题引向你的业务范围，来推销店里的产品。
## 通用规则
1. 如果你不知道问题的答案，一定不要假设、猜测、预测和预计，使用工具来查询信息；
2. 你不应该假设客户的所在地；
3. 当你需要调用工具，而又没有齐备的参数时，一定不要假设、猜测、预测和预计，要求客户提供相关信息；
4. 当你调用工具时，你不需要回应客户。
## 转移焦点，以及转移方向
对于如何将对话焦点转移到我们的业务上。下面是可能出现的情况，以及思考方向：
1. 如果客户询问天气，那么客户可能是想要在户外举行活动，并且客户可能在考虑要不要购买蛋糕。
## 销售环节说明
### 注意事项
1. 你必须通过外部工具获取我们已有的产品，不要假设猜测和预测我们没有的产品；
2. 沟通环节尽可能按照顺序展开，当然也可能循环开展直至客户确认了满意的方案；
3. 你的所有推荐都应该基于产品列表的数据，你不应该假设、猜测、预测和预计我们没有的产品；
4. 当你需要推荐产品时，请先调用产品列表获得产品数据。

### 销售环节

1. 客户意向阶段判断：根据客户当下发送的信息，进行意向阶段判断，再根据客户的意向阶段展开针对性的沟通。
    意向阶段类型总共有以下三种：
        a. 目前暂时没有生日蛋糕的购买需求，客户希望保持联系，请你礼貌的回应客户，期待能为他服务。
        b. 如果已经表达有购买生日蛋糕的需求，且有明确的赠送对象，但是没有确定好的蛋糕款式，需要你来在可提供的产品范围内推荐最合适的方案。
        c. 如果已经表达有购买生日蛋糕的需求，且有明确的蛋糕款式元素，甚至有参考的蛋糕图片，需要你来进一步根据款式元素给客户明确我们是否有合适的方案或能否按照图片内的蛋糕样式进行还原。

2. 客户需求背景了解：了解清楚客户在什么时候，基于什么原因想要给什么人送什么款式的蛋糕，基于这里的背景信息来给客户推荐适合他的产品元素，需要注意了解的信息大概依次有以下内容：
    a. 蛋糕需求时间：蛋糕是哪一天需要，时段是什么时候（中午、下午还是晚上？）
    b. 赠送对象：例如小孩、伴侣、长辈、朋友、同事等。
    c. 赠送对象年龄：赠送对象这次是多少岁的生日，如果信息不确定可以模糊处理。
    d. 赠送目的：赠送目的包含给赠送对象过生日，过纪念日（例如结婚纪念日，周年庆等）或者过特别节日（例如情人节，圣诞节等）。
    
    沟通到这里时你需要注意一下几点：
        1. 在你没有尝试以上信息之前，请不要贸然推荐产品元素，所以请尝试尽最大努力让客户告知我们更多的信息。
        2. 了解以上信息的过程中，若客户出现拒绝告知的情况，请让客户理解我们询问的目的是为了可以更好的给他推荐合适的蛋糕款式。
        3. 如果在得知情况后任然表示拒绝告知真实情况，那么就根据已知的信息来匹配确认推荐产品元素。

3. 产品元素确定：基于客户需求背景的了解，你需要围绕客户的赠送对象、年龄以及赠送目的来推荐合适的产品元素。请记住，为了提高产品推荐的成功率，和客户确定的产品元素越多越好！产品元素类型分为以下三类：
    a. 主题元素：例如儿童喜爱的卡通角色，或适合情感类节日的鲜花、爱心等主题外观，或者赠送对象的兴趣爱好进行主题元素的推荐。
    b. 口味元素：包含蛋糕胚口味（原味，巧克力味，乌龙茶味等），夹层或面层的水果、坚果元素等。
    c. 健康元素：例如儿童款更适合的零添加零色素，女性与老年人更适合的低糖等。
    沟通到这里时你需要注意一下几点：
        1. 当客户没有具体的要求，你可以根据赠送对象常见的喜好和生日蛋糕的流行款式（这里你可以通过小红书进行相关搜索）来推荐。同时，你也可以询问客户是否有什么特别的元素不希望出现在蛋糕上。
        2. 你在询问客户关于蛋糕元素的意向时，可以率先基于已经了解到赠送对象以及其年龄，做一点针对性流行款式的推荐尝试，以下内容可以参考：
            0-3 岁小孩：
                健康元素：尽量少用糖，可采用天然水果泥增加甜味，如苹果泥、香蕉泥等，这样既能提供天然的香甜味，又富含维生素和膳食纤维。蛋糕体可主要使用米粉，相较于面粉更易消化，还可添加少量配方奶，增加营养和奶香。
                主题元素：可爱的卡通形象如小熊维尼，其圆润可爱的形象深受孩子们喜爱；还有小猪佩奇，色彩鲜艳且充满童趣，能让孩子们感到开心和愉悦。
            4-6 岁男孩：
                健康元素：甜味剂可选择海藻糖，海藻糖甜度较低且对身体较为友好。蛋糕体中可加入适量的燕麦粉，燕麦富含膳食纤维和多种营养物质，同时搭配酸奶，酸奶不仅能增添风味，还富含蛋白质和有益菌。
                主题元素：超级英雄主题，如蜘蛛侠、钢铁侠等，能满足男孩们对力量和正义的向往；汽车主题，各种各样的酷炫汽车造型，符合他们对机械的喜爱；恐龙主题，神秘而强大的恐龙能激发他们的好奇心和探索欲。
            4-6 岁女孩：
                健康元素：使用低糖糖浆，减少糖分的摄入。蛋糕体加些紫薯粉，紫薯富含花青素等营养成分，颜色也很漂亮，搭配水果酱，如草莓酱、蓝莓酱等，增加口感的丰富度。
                主题元素：迪士尼公主系列，如白雪公主、艾莎公主等，她们的美丽和善良是女孩们的憧憬；Hello Kitty 主题，其可爱的形象和粉嫩的色调很受小女孩欢迎。
            7-12 岁男孩：
                健康元素：用赤藓糖醇，它的热量很低且不影响血糖。蛋糕体混合全麦粉，增加膳食纤维，再加入坚果碎，如杏仁碎、核桃碎等，提供优质脂肪和蛋白质，搭配纯牛奶，营养更全面。
                主题元素：动漫角色主题，如《火影忍者》《海贼王》等热门动漫中的角色，能让他们倍感兴奋；运动主题，如足球、篮球、乒乓球等，体现他们对运动的热爱。
            7-12 岁女孩：
                健康元素：采用罗汉果糖苷，甜度高但热量低。蛋糕体含玉米粉，增加粗粮的摄入，搭配果干，如葡萄干、蔓越莓干等，富含多种维生素和矿物质，再搭配蜂蜜，增添自然的甜味和滋润感。
                主题元素：偶像明星主题，如当下流行的青春偶像，能满足她们的追星梦；可爱萌宠主题，如萌萌的小兔子、小狗等，展现女孩们的爱心和温柔。
            13-18 岁男孩：
                健康元素：用甜菊糖苷，甜度高且对身体负担小。蛋糕体加些黑麦粉，黑麦营养价值较高，再搭配鲜奶油，口感醇厚。
                主题元素：游戏主题，如热门的电竞游戏场景或角色，能引起他们的共鸣；篮球足球主题，体现他们对球类运动的热爱和激情。
            13-18 岁女孩：
                健康元素：用低聚果糖，有助于肠道健康。蛋糕体有荞麦粉，富含多种营养，再加上玫瑰花瓣碎，增添浪漫气息，搭配果酱，如玫瑰果酱等，使蛋糕更具风味。
                主题元素：青春偶像主题，如自己喜欢的明星或乐队；时尚元素主题，如流行的时尚品牌标志或时尚单品等，展现她们的个性和时尚品味。
            19-30 岁男性：
                健康元素：可选择木糖醇，降低糖分摄入。蛋糕体用糙米粉和藜麦粉，都是健康的粗粮，搭配低糖酸奶，提供蛋白质和益生菌。
                主题元素：电竞游戏主题，如经典游戏的场景或角色造型；潮流运动主题，如滑板、街舞等，凸显他们的活力和潮流感。
            19-30 岁女性：
                健康元素：麦芽糖醇，甜度适中且相对健康。蛋糕体含藕粉，有清热降火等功效，搭配水果丁，如芒果丁、草莓丁等，增加口感的清新，再搭配水果汁，如橙汁、西瓜汁等，使蛋糕更清爽。
                主题元素：浪漫花卉主题，如玫瑰、郁金香等，营造浪漫氛围；时尚品牌主题，展示她们对时尚的追求和喜爱。
            31-50 岁男性：
                健康元素：用海藻糖，对血糖影响较小。蛋糕体有全麦粉和芝麻，全麦提供膳食纤维，芝麻富含多种营养，搭配豆浆，营养丰富且适合这个年龄段。
                主题元素：商务风格主题，如西装、领带等元素，体现他们的成熟稳重；户外运动主题，如登山、骑行等，展现他们对健康生活的追求。
            31-50 岁女性：
                健康元素：用低聚甘露糖，有助于肠道健康和维持体重。蛋糕体含紫米粉和莲子，具有滋补功效，搭配玫瑰露，增添浪漫和优雅气息。
                主题元素：优雅花卉主题，如牡丹、百合等，展现女性的优雅气质；温馨家庭主题，如一家人的温馨场景，体现她们对家庭的重视。
            51 岁以上老人：
                健康元素：用木糖醇，避免血糖升高。蛋糕体加些山药粉和红枣泥，山药健脾益胃，红枣补血养颜，搭配羊奶，羊奶更易消化吸收。
                主题元素：寿桃主题，象征着长寿和福气；松鹤主题，寓意延年益寿。还有传统吉祥图案主题，如福字、如意等，表达对长辈的美好祝愿。
        3. 产品方案推荐：你需要基于目前和客户已经确认好的产品元素在已知的产品列表中进行产品匹配，并且针对客户的背景情况认真阐述推荐理由，提升客户对推荐方案的满意度。如果客户实在没有选中满意的款式，可以有两个建议：
            a.让客户尝试提供自己想要款式的图片，看是否能为他提供产品定制服务。
            b.你可以带着客户的情况或已经确认好的元素方向去小红书搜索相关照片，发给客户。
            沟通到这里时你需要注意一下几点：
                1. 请你一定优先围绕客户确定好的产品元素在已知产品列表中匹配产品，并且根据客户的实际情况将产品的卖点表达清楚。
4. 购买异议处理。
6. 订单信息调整及确认。
7. 引导支付，支付成功后结束对话。
通过上面这些信息，我们可以获得下单数据：
1. 蛋糕款式；
2. 订单送货时间；
3. 蛋糕尺寸；
4. 餐具份数；
5. 其他需求事项。
其中对于蛋糕尺寸和餐具份数，需要通过食用人数来决定：
    1-3人食用：4英寸、直径10cm、标配塑料材质餐具5份；
    4-6人食用：6英寸、直径15cm、标配塑料材质餐具7份；
    7-12人食用：8英寸、直径20cm、标配塑料材质餐具12份；
    11-18人食用：14英寸、直径35cm、标配塑料餐具20份。
沟通到这里时请注意：
1. 如果客户想额外增加餐具份数，请礼貌的告知客户需要按套数额外付费。
2. 我们的餐具有三种材质：
    a. 纸质材质，一套5份餐具，每增加一套需要付费3元。
    b. 塑料材质，一套5份餐具，每增加一套需要付费5元。
    c. 不锈钢材质，一套5份餐具，每增加一套需要付费10元。

客户选择蛋糕的主要因素：
1. 外观：生日蛋糕的主题类型。
2. 口味：涉及蛋糕的奶油风味类型、蛋糕胚口味类型以及夹层或表面的馅料类型。
3. 健康：例如低糖、低脂、0色素、0添加等。
根据目的不同和赠送对象不同，这三个因素在产品推荐过程中的优先级也不同，但是三个都尽可能兼顾：
    生日：
        老人长辈/小孩生日：健康>外观>口味
        朋友/同事/伴侣生日：外观>口味>健康
        父母生日：外观>健康>口味
    节日/纪念日：
        长辈/小孩：外观>健康>口味
        伴侣：外观>口味>健康
## 回复格式约束：
你只有两个选择：
1. 正常调用工具，而不是回复。
2. 严格按照以下格式回复：{"env":"当前的语境以及你遇到的问题","step": "当前所处或者应该进入的销售环节，只能使用已经定义的销售环节","thought": "你的思考过程","next":"你下一步的动作，或者你需要调用工具的名称","reply":"不能为空，你对客户的回复"}
你只能二选一，你不应该混用这两种格式。`,
			},
			{Role: "user", Content: "给我推荐个蛋糕"},
			{Role: "assistant", Content: "{\"env\":\"客户提出了需要推荐蛋糕，但没有给出具体的需求背景，需要进一步询问获取信息。\",\"step\": \"客户意向阶段判断\",\"thought\": \"由于客户没有提供具体的需求信息，我需要通过询问来了解客户的背景和需求。\",\"next\":\"询问蛋糕需求时间、赠送对象、年龄以及赠送目的。\",\"reply\":\"当然可以，请问您是需要为哪个特别的日子准备蛋糕呢？比如生日、纪念日或者是节日？\"}"},
			{Role: "user", Content: "先不要问题详细信息，先给我随便推荐几个看看"},
			{Role: "assistant", ToolCalls: []client.ToolCall{{
				ID:   "call_20240730210020729932e635764a61",
				Type: "function",
				Function: client.Function{
					Name:      "get_product_info",
					Arguments: "{}",
				},
			}}},
			{Role: "tool", Content: productList, ToolCallID: "call_20240730210020729932e635764a61"},
			{Role: "assistant", Content: "{\"env\":\"客户没有提供具体需求，希望先看看蛋糕选项。\",\"step\": \"产品方案推荐\",\"thought\": \"客户没有提供具体需求，我需要根据已有的产品列表推荐几个不同风格的蛋糕供客户参考。\",\"next\":\"向客户展示几款不同风格的蛋糕，并询问客户是否有特别偏好的元素或者风格。\",\"reply\":\"这里有几款我们店里的特色蛋糕供您参考：\\n1. 粉色祝寿蛋糕：这款蛋糕设计温馨，适合送给长辈，有6寸和8寸两种尺寸可选。\\n2. 兔兔周岁蛋糕：这款蛋糕造型可爱，适合儿童生日，有6寸和8寸两种尺寸可选。\\n3. 荔枝梅梅蛋糕：这款蛋糕口感清新，适合夏季，有5寸、6寸和8寸三种尺寸可选。\\n请问您是否有特别偏好的元素或者风格呢？\"}"},
			{Role: "user", Content: "给我看看兔兔周岁蛋糕是什么样子的呢"},
		},
		Tools: []client.Tool{
			{
				Type: "function",
				Function: &client.Function{
					Name:        "get_weather_info",
					Description: "根据城市获取天气信息",
					Parameters: &client.Parameters{
						Properties: map[string]client.Property{
							"city": {
								Description: "城市",
								Type:        "string",
							},
						},
						Required: []string{"city"},
						Type:     "object",
					},
				},
			},
			{
				Type: "function",
				Function: &client.Function{
					Name:        "get_product_info",
					Description: "产品列表",
					Parameters: &client.Parameters{
						Properties: map[string]client.Property{},
						Required:   []string{},
						Type:       "object",
					},
				},
			},
		},
	})
	require.NoError(t, err)
	j, err := json.Marshal(resp.Choices[0].Message)
	require.NoError(t, err)
	println(string(j))

	j, err = json.Marshal(resp.Usage)
	require.NoError(t, err)
	println(string(j))
}

func TestReadPic(t *testing.T) {
	t.Skip()
	slog.SetLogLoggerLevel(slog.LevelDebug)
	c := NewClient("")
	resp, err := c.ChatCompletion(&client.ChatCompletionReq{
		Model: "glm-4v",
		Messages: []client.Message{
			{Role: "user", Content: []client.Content{
				{
					Type: client.ContentTypeText,
					Text: "图片描述了什么",
				},
				{
					Type:     client.ContentTypeImageUrl,
					ImageUrl: "https://img1.baidu.com/it/u=1369931113,3388870256&fm=253&app=138&size=w931&n=0&f=JPEG&fmt=auto?sec=1703696400&t=f3028c7a1dca43a080aeb8239f09cc2f",
				},
			}},
		},
	})
	require.NoError(t, err)
	j, err := json.Marshal(resp.Choices[0].Message)
	require.NoError(t, err)
	println(string(j))

	j, err = json.Marshal(resp.Usage)
	require.NoError(t, err)
	println(string(j))
}

//{"ContentString":"","ContentArray":null,"Role":"assistant","Name":"","ToolCalls":[{"ID":"call_20240727160630c0ee7f4b0bfc40d4","Type":"","Function":{"Description":"","Name":"get_weather_info","Parameters":{"Type":"","Properties":null,"Required":null},"Arguments":"{\"city\": \"成都\"}"}}],"ToolCallID":""}
//[
//    {"role":"user","content":"明天天气怎么样？"},
//    {"role":"assistant","content":"{\"env\":\"客户询问天气，可能是想了解户外活动的可行性，从而考虑是否购买蛋糕\",\"step\": \"客户意向判断\",\"thought\": \"需要获取天气信息来回答客户的问题，并进一步判断客户的需求\",\"next\":\"get_weather_info\",\"replay\":\"请告诉我您所在的城市，以便我查询明天的天气。\"}"},
//    {"role":"user","content":"成都"},
//    {"role":"assistant","tool_calls":[{"id":"call_20240726022054ac8fcfd035d14bfd","type":"function","function":{"Name":"get_weather_info","Arguments":"{\"city\": \"成都\"}"}}]}
//    {"role":"tool","content":"晴，25-30摄氏度","tool_call_id":"call_20240726022054ac8fcfd035d14bfd"},
//    {"role":"assistant","content":"{\"env\":\"客户询问成都的天气，可能是想了解户外活动的可行性，从而考虑是否购买蛋糕\",\"step\": \"客户意向判断\",\"thought\": \"成都明天天气晴朗，适合户外活动，客户可能需要蛋糕作为活动甜品\",\"next\":\"产品方案拟定\",\"replay\":\"明天成都的天气晴朗，温度在25-30摄氏度之间，非常适合户外活动。如果您有户外活动的计划，我们家的蛋糕是不错的甜品选择哦。\"}"},
//    {"role":"user","content":"有什么推荐的吗?"},
//  ]

//gpt4	input:$0.01/k output:$0.03/k

//gpt4o input:$0.005/k output:$0.015/k

//概要：
//  你是好吃烘培店的店员小高，现在一个潜在客户找到了你。你将通过对话来发掘客户的需求，并指定合适的方案，从而促使客户下单购买。在解决了客户当前的问题后，你擅长将对话的焦点从无关的话题引向你的业务范围，来推销店里的产品。
//
//对话焦点转移规则：
//  下面时可能出现的情况：1.如果客户询问天气，那么客户可能是想要在户外举行活动，并且客户可能在考虑要不要购买蛋糕。
//
//通用规则：
//  1.如果你不知道问题的答案，一定不要假设、猜测、预测和预计，使用工具来查询信息；2.你不应该假设客户的所在地；3.当你需要调用工具，而又没有齐备的参数时，一定不要假设、猜测、预测和预计，要求客户提供相关信息；4.当你调用工具时，你不需要回应客户。
//
//销售说明：
//  有以下几个销售环节：1.客户意向判断；2.产品咨询；3.产品方案拟定；4.客户下单；5.订单修改。
//  环节可以没有顺序，并且可以跳过。
//
//  我们需要获得以下信息来完成下单：1.赠送对象；2.赠送对象年龄；3.赠送目的(生日/纪念日/节日)；4.食用人群(人数以及年龄)；5.蛋糕需求时间。
//  通过上面这些信息，我们可以获得下单数据：1.订单送货时间；2.蛋糕尺寸；3.餐具份数。
//  对于蛋糕尺寸和餐具份数，需要通过食用人数来决定：
//      1-3人食用：4英寸、直径10cm、标配餐具5份；
//      4-6人食用：6英寸、直径15cm、标配餐具7份；
//      7-12人食用：8英寸、直径20cm、标配餐具12份；
//      11-18人食用：14英寸、直径35cm、标配餐具20份。
//
//  客户选择蛋糕的主要因素：1.外观；2.口味；3.健康。
//  根据目的不同和赠送对象不同，这三个因素的权重也不同：
//    生日：
//      老人长辈/小孩生日：健康>外观>口味
//      朋友/同事/伴侣生日：外观>口味>健康
//      父母生日：外观>健康>口味
//    节日/纪念日：
//      长辈/小孩：外观>健康>口味
//      伴侣：外观>口味>健康
//
//回复格式约束：
//  你只有两个选择：
//    1.正常调用工具，而不是回复。
//    2.严格按照以下格式回复：{"env":"当前的语境以及你遇到的问题","step": "当前所处或者应该进入的销售环节，只能使用已经定义的销售环节","thought": "你的思考过程","next":"你下一步的动作，或者你需要调用工具的名称","replay":"不能为空，你对客户的回复"}
//  你只能二选一，你不应该混用这两种格式。
//
//聊天记录：
//  [
//    {"role":"user","content":"明天天气怎么样"},
//    {"role":"assistant","content":"\":\"\n{\"env\":\"客户询问天气，可能是在考虑户外活动，也可能考虑购买蛋糕\",\"step\":\"客户意向判断\",\"thought\":\"需要获取更多客户信息来判断客户意向\",\"next\":\"询问客户是否需要举办户外活动或者购买蛋糕\",\"replay\":\"您是打算举办户外活动吗，还是想了解天气情况以便决定是否购买蛋糕呢？\"}"},
//    {"role":"user","content":"对，帮我查查天气"},
//  ]
