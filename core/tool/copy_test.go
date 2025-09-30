package tool

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/require"
)

func TestSimpleMap(t *testing.T) {
	t.Run("entity_to_model", func(t *testing.T) {
		type CommonField struct {
			ID uint64
		}

		type UserEntity struct {
			Id       uint64
			Username string
		}

		type UserModel struct {
			CommonField
			UserName string
		}

		e := UserEntity{
			Id:       1,
			Username: "zed",
		}
		var m UserModel
		err := SimpleMap(&m, e)
		require.NoError(t, err)
		require.Equal(t, e.Username, m.UserName)
		require.Equal(t, e.Id, m.ID)
	})
	t.Run("model_to_entity", func(t *testing.T) {
		type CommonField struct {
			ID uint64
		}

		type UserEntity struct {
			Id       uint64
			Username string
		}

		type UserModel struct {
			CommonField
			UserName string
		}

		m := UserModel{
			CommonField: CommonField{
				ID: 1,
			},
			UserName: "zed",
		}
		var e UserEntity
		err := SimpleMap(&e, m)
		require.NoError(t, err)
		require.Equal(t, m.UserName, e.Username)
		require.Equal(t, m.ID, e.Id)
	})
	t.Run("anonymous", func(t *testing.T) {
		type CommonField struct {
			ID uint64
		}

		type UserEntity struct {
			CommonField
			Username string
		}

		type UserModel struct {
			CommonField
			UserName string
		}

		m := UserModel{
			CommonField: CommonField{
				ID: 1,
			},
			UserName: "zed",
		}
		var e UserEntity
		err := SimpleMap(&e, m)
		require.NoError(t, err)
		require.Equal(t, m.UserName, e.Username)
		require.Equal(t, m.ID, e.ID)
	})
	t.Run("TestFieldInValid", func(t *testing.T) {
		type CommonField struct {
			C bool
		}

		type s1 struct {
			A int
			B string
			CommonField
		}

		type s2 struct {
			A int
			B string
			CommonField
		}

		struct1 := s1{
			A:           1,
			B:           "2",
			CommonField: CommonField{C: true},
		}
		var struct2 s2
		err := SimpleMap(&struct2, struct1)
		if err != nil {
			require.NoError(t, err)
		}
		require.Equal(t, struct1.A, struct2.A)
		require.Equal(t, struct1.B, struct2.B)
		require.Equal(t, struct1.C, struct2.C)
	})
	t.Run("TestSimpleMap2", func(t *testing.T) {
		type m struct {
			ID       uint64
			Username string
			Password string
		}
		type e struct {
			ID       uint64
			Username string
			Password string
		}

		model := &m{
			ID:       1,
			Username: "zed",
			Password: "zed",
		}
		var entity e
		err := SimpleMap(&entity, &model)
		require.NoError(t, err)
		require.Equal(t, model.ID, entity.ID)
		require.Equal(t, model.Username, entity.Username)
		require.Equal(t, model.Password, entity.Password)
	})
	t.Run("TestSimpleMap3", func(t *testing.T) {
		type Common1 struct {
			C bool
		}
		type Common2 struct {
			C bool
		}
		type s1 struct {
			A      int
			Common *Common1
		}
		type s2 struct {
			A      int
			Common Common2
		}

		struct1 := s1{
			A:      1,
			Common: &Common1{C: true},
		}
		var struct2 s2
		err := SimpleMap(&struct2, struct1)
		require.NoError(t, err)
		require.Equal(t, struct1.A, struct2.A)
		require.Equal(t, struct1.Common.C, struct2.Common.C)
	})
	t.Run("TestSimpleMap4", func(t *testing.T) {
		type S1 struct {
			A string
		}
		type S2 struct {
			A []string
		}
		s1 := S1{}
		var s2 S2
		err := SimpleMap(&s2, s1)
		require.NoError(t, err)
		require.Equal(t, 0, len(s2.A))
	})
	t.Run("TestSimpleMap5", func(t *testing.T) {
		type Common1 struct {
			C bool
		}
		type s1 struct {
			A int
			*Common1
		}
		type s2 struct {
			A int
			*Common1
		}

		struct1 := s1{
			A:       1,
			Common1: &Common1{C: true},
		}
		var struct2 s2
		err := SimpleMap(&struct2, struct1)
		require.NoError(t, err)
		require.Equal(t, struct1.A, struct2.A)
		require.Equal(t, struct1.Common1.C, struct2.Common1.C)
	})
	t.Run("TestSimpleMap6", func(t *testing.T) {
		type Common1 struct {
			C bool
		}
		type Common2 struct {
			C bool
		}
		type s1 struct {
			A      int
			Common *Common1
		}
		type s2 struct {
			A      int
			Common *Common2
		}

		struct1 := s1{
			A:      1,
			Common: &Common1{C: true},
		}
		var struct2 s2
		err := SimpleMap(&struct2, struct1)
		require.NoError(t, err)
		require.Equal(t, struct1.A, struct2.A)
		require.Equal(t, struct1.Common.C, struct2.Common.C)
	})
	t.Run("TestSimpleMap7", func(t *testing.T) {
		type c1 struct {
			B int
		}
		type c2 struct {
			B int
		}
		type s1 struct {
			Bs []*c1
		}
		type s2 struct {
			Bs []*c2
		}

		struct1 := s1{
			Bs: []*c1{{B: 1}},
		}
		var struct2 s2
		err := SimpleMap(&struct2, struct1)
		require.NoError(t, err)
		require.Equal(t, struct1.Bs[0].B, struct2.Bs[0].B)
	})
	t.Run("TestSimpleMap8", func(t *testing.T) {
		type struct1 struct {
			A map[string]string
		}
		type struct2 struct {
			A map[string]string
		}
		s1 := struct1{A: map[string]string{"test": "1"}}
		var s2 *struct2
		err := SimpleMap(&s2, s1)
		require.NoError(t, err)
		require.Equal(t, s1.A["test"], s2.A["test"])
	})
	t.Run("TestSimpleMap9", func(t *testing.T) {
		type struct1 struct {
			A map[string]string
		}
		type struct2 struct {
			A map[string]string
		}
		var s1 *struct1
		var s2 *struct2
		err := SimpleMap(&s2, s1)
		require.NoError(t, err)
		//require.Equal(t, s1.A["test"], s2.A["test"])
	})
	t.Run("TestSimpleMapSlice1", func(t *testing.T) {
		type s1 struct {
			A int
		}
		type s2 struct {
			A int
		}

		slice1 := []s1{
			{A: 1},
		}
		slice2 := make([]s2, 0)
		err := SimpleMap(&slice2, slice1)
		require.NoError(t, err)
		require.Equal(t, slice1[0].A, slice2[0].A)
	})
	t.Run("TestSimpleMapSlice2", func(t *testing.T) {
		type s1 struct {
			A int
		}
		type s2 struct {
			A int
		}

		slice1 := []s1{
			{A: 1},
		}
		var slice2 []s2
		err := SimpleMap(&slice2, slice1)
		require.NoError(t, err)
		require.Equal(t, slice1[0].A, slice2[0].A)
	})
	t.Run("TestSimpleMapSlice3", func(t *testing.T) {
		type struct1 struct {
			A bool
		}
		type struct2 struct {
			A bool
		}
		s1 := []struct1{{A: true}}
		var s2 []struct2
		err := SimpleMap(&s2, s1)
		require.NoError(t, err)
		require.Equal(t, s1[0].A, s2[0].A)
	})
	t.Run("TestSimpleMapSlice4", func(t *testing.T) {
		type Common2 struct {
			ID uint64 `json:"id,string"`
		}

		type struct2 struct {
			Types         []int    `json:"types"`
			CurrentNumber *Common2 `json:"currentNumber"`
		}

		type Common1 struct {
			ID uint64
		}

		type struct1 struct {
			Types         []int
			CurrentNumber *Common1
		}

		s1 := []struct1{{Types: []int{}, CurrentNumber: &Common1{ID: 1}}}
		var s2 []*struct2
		err := SimpleMap(&s2, s1)
		require.NoError(t, err)
		require.Equal(t, uint64(1), s2[0].CurrentNumber.ID)
	})
	t.Run("TestSimpleMapUnAddr", func(t *testing.T) {
		//known issue
		require.Panics(t, func() {
			type struct2 struct {
				A int
			}
			type struct1 struct {
				A int
			}
			s1 := struct1{A: 1}
			var s2 *struct2
			err := SimpleMap(s2, s1)
			require.Error(t, err)
		})
	})
	t.Run("空字符串转json null", func(t *testing.T) {
		type c struct {
			C string
		}
		type a struct {
			C *c
		}
		type b struct {
			C string
		}
		aa := a{}
		var bb b
		err := SimpleMap(&bb, &aa)
		require.NoError(t, err)
		require.Equal(t, "null", bb.C)
	})
	t.Run("[]byte类型json转struct", func(t *testing.T) {
		type c struct {
			C string `json:"c"`
		}
		type a struct {
			C *c
		}
		type b struct {
			C []byte
		}
		bb := b{
			C: []byte(`{"c":"123"}`),
		}
		var aa a
		err := SimpleMap(&aa, bb)
		require.NoError(t, err)
		require.Equal(t, "123", aa.C.C)
	})
	t.Run("[]byte类型json,struct互转", func(t *testing.T) {
		type c struct {
			C string `json:"c"`
		}
		type a struct {
			C c
		}
		type b struct {
			C []byte
		}
		aa := a{
			C: c{
				C: "123",
			},
		}
		var bb b
		err := SimpleMap(&bb, aa)
		require.NoError(t, err)
		err = SimpleMap(&aa, bb)
		require.NoError(t, err)
		require.Equal(t, "123", aa.C.C)
	})
	t.Run("[]byte,struct互转", func(t *testing.T) {
		a := PlanGroupEnd{
			ID:         1,
			MerchantID: 1,
			Name:       "12321testname",
			TaskCount:  1,
			ConfigJSON: PlanGroupConfigJSON{
				OpportunityIDs:   []uint64{1, 2, 3},
				OpportunityNames: []string{"123", "456", "789"},
				IsStop:           true,
				Tasks: []TaskInfo{
					{
						AdvanceDays: 5,
						SendTime:    "10:00",
						ContentList: []ContentItem{
							{
								Type:       1,
								MaterialID: 1,
								Content:    "123",
							},
						},
					},
				},
			},
			FollowUpConfigJSON: FollowUpConfigJSON{
				CouponConfig: &CouponFollowUpConfig{},
				CartConfig:   &CartFollowUpConfig{},
				DemandConfig: &DemandFollowUpConfig{},
			},
			Plans: []*Plan{},
		}
		b := PlanGroup{}
		err := SimpleMap(&b, a)
		require.NoError(t, err)
		err = SimpleMap(&a, b)
		require.NoError(t, err)
		fmt.Printf("%+v\n", b)
	})
	t.Run("逗号隔开的字符串转slice", func(t *testing.T) {
		type struct1 struct {
			A string
		}
		type struct2 struct {
			A []uint64
		}
		s1 := struct1{A: "1"}
		var s2 struct2
		err := SimpleMap(&s2, s1)
		require.NoError(t, err)
		require.Equal(t, []uint64{1}, s2.A)
	})
	t.Run("逗号隔开的字符串转slice2", func(t *testing.T) {
		type struct1 struct {
			A string
		}
		type struct2 struct {
			A []string
		}
		s1 := struct1{A: "呵呵,123"}
		var s2 struct2
		err := SimpleMap(&s2, s1)
		require.NoError(t, err)
		require.Equal(t, []string{"呵呵", "123"}, s2.A)
	})
}

type PlanGroupEnd struct {
	ID                 uint64              `json:"id"`
	MerchantID         uint64              `json:"merchant_id"` // 商户ID
	Name               string              `json:"name"`
	OpportunityCount   uint                `json:"opportunity_count"`
	TaskCount          uint                `json:"task_count"`
	ConfigJSON         PlanGroupConfigJSON `json:"config_json"`
	FollowUpConfigJSON FollowUpConfigJSON  `json:"follow_up_config_json"` // 追单配置JSON
	CreatedAt          int64               `json:"created_at"`
	UpdatedAt          int64               `json:"updated_at"`
	DeletedAt          int64               `json:"deleted_at"`

	// 关联字段
	Plans []*Plan `json:"plans,omitempty"`

	// 统计字段
	TotalTaskCount   uint32 `json:"total_task_count"`   // 总任务数
	TotalTargetCount uint32 `json:"total_target_count"` // 总目标数
	TotalArriveCount uint32 `json:"total_arrive_count"` // 总成功抵达数
	TotalReplyCount  uint32 `json:"total_reply_count"`  // 总回复数
}

type ConfigJSON struct {
	CommStyle             string `json:"comm_style"`               // 沟通风格
	OtherDemands          string `json:"other_demands"`            // 其他诉求
	FirstTouchAdvanceDays int32  `json:"first_touch_advance_days"` // 第一次触达提前天数
}

type PreferenceJSON struct {
	Theme  string `json:"theme"`  // 主题
	Taste  string `json:"taste"`  // 口味
	Health string `json:"health"` // 健康
}

type Plan struct {
	ID                 uint64             `json:"id"`
	MerchantID         uint64             `json:"merchant_id"` // 商户ID
	PlanGroupID        uint64             `json:"plan_group_id"`
	OpportunityID      uint64             `json:"opportunity_id"`
	WechatContactID    uint64             `json:"wechat_contact_id"` // 微信联系人ID
	Name               string             `json:"name"`
	NickName           string             `json:"nick_name"` // 昵称
	TaskCount          uint               `json:"task_count"`
	ConfigJSON         ConfigJSON         `json:"config_json"`
	FollowUpConfigJSON FollowUpConfigJSON `json:"follow_up_config_json"` // 追单配置JSON
	ExpectTouchedAt    int64              `json:"expect_touched_at"`
	HasOrder           uint8              `json:"has_order"` // 是否有订单：0-无 1-有
	Relation           string             `json:"relation"`
	DemandScenario     string             `json:"demand_scenario"`
	CalendarSystem     string             `json:"calendar_system"`
	Age                uint               `json:"age"`
	Status             uint8              `json:"status"`             // 状态：1-待执行 2-执行中 3-已完成 4-已终止
	IsStop             uint8              `json:"is_stop"`            // 是否停止后续节点：0-否 1-是
	IsAIGenerated      uint8              `json:"is_ai_generated"`    // 是否由AI生成：0-否 1-是
	AIWorkflowStatus   int8               `json:"ai_workflow_status"` // AI工作流状态：0-未开始 1-生成中 2-生成完成 3-生成失败
	ResponseJSON       string             `json:"response_json"`      // AI响应JSON字符串
	PreferenceJSON     PreferenceJSON     `json:"preference_json"`
	ArriveCount        uint32             `json:"arrive_count"` // 成功触达数
	ReplyCount         uint32             `json:"reply_count"`  // 回复数
	CreatedAt          int64              `json:"created_at"`
	UpdatedAt          int64              `json:"updated_at"`
	DeletedAt          int64              `json:"deleted_at"`
	Tasks              []*Task            `json:"-" gorm:"-"` // 关联的任务列表，不进行JSON序列化和GORM映射
	// OpportunityExists 表示该计划关联的商家是否存在,**只在列表请求中有意义**.
	//  因重新分析导致的商机变动,关联的商机可能会不存在,故在列表请求中检查.
	//  重新分析需要变动商机的理由是: 重新分析表示之前的分析可能存在错误,故相关的商机会被删除后重建,继而id变更.
	OpportunityExists bool  `json:"opportunity_exists"`
	ReplyTime         int64 `json:"-" gorm:"-"`
	ArriveTime        int64 `json:"-" gorm:"-"`
}

type Task struct {
	ID                      uint64                  `json:"id"`
	PlanID                  uint64                  `json:"plan_id"`
	WechatContactID         uint64                  `json:"wechat_contact_id"` // 微信联系人ID
	MerchantID              uint64                  `json:"merchant_id"`
	ExecuteID               uint64                  `json:"execute_id"` // 执行ID
	Title                   string                  `json:"title"`
	TaskType                uint8                   `json:"task_type"` // 任务类型：1-普通类型 2-优惠券追单类型 3-购物车追单类型 4-需求时间追单类型
	Status                  int8                    `json:"status"`
	Reason                  string                  `json:"reason"`
	ExecuteTime             int64                   `json:"execute_time"`
	ReplyTime               int64                   `json:"reply_time"`
	ArriveTime              int64                   `json:"arrive_time"`
	ContentJSON             TaskContent             `json:"content_json"`
	FollowUpTriggerDataJSON TaskFollowUpTriggerData `json:"follow_up_trigger_data_json"` // 追单触发数据JSON
	TargetName              string                  `json:"target_name"`
	AdvanceDays             int                     `json:"advance_days"` // 提前天数
	SendTime                string                  `json:"send_time"`    // 发送时间，格式：HH:mm:ss
	CreatedAt               int64                   `json:"created_at"`
	UpdatedAt               int64                   `json:"updated_at"`
	DeletedAt               int64                   `json:"deleted_at"`
	ReplyCount              uint32                  `json:"reply_count"` // 回复数
	PlanDemandScenario      string                  `json:"plan_demand_scenario"`
	ContentTemplateID       uint64                  `json:"content_template_id"` // 内容模板ID
}

type TaskFollowUpTriggerData struct {
	CouponID     uint64 `json:"coupon_id"`     // 优惠券ID（优惠券追单类型使用）
	CouponName   string `json:"coupon_name"`   // 优惠券名称
	CartData     string `json:"cart_data"`     // 购物车数据（购物车追单类型使用）
	DemandTime   int64  `json:"demand_time"`   // 需求时间（需求时间追单类型使用）
	TriggerTime  int64  `json:"trigger_time"`  // 触发时间
	CustomerID   uint64 `json:"customer_id"`   // 客户ID
	CustomerName string `json:"customer_name"` // 客户名称
}

type TaskContent struct {
	ContentList []ContentItem `json:"content_list"` // 内容列表
}

type PlanGroupConfigJSON struct {
	ProductIDs            []uint64   `json:"product_ids"`              // 商品ID列表
	ProductNames          []string   `json:"product_names"`            // 商品名称列表
	OpportunityIDs        []uint64   `json:"opportunity_ids"`          // 商机ID列表
	OpportunityNames      []string   `json:"opportunity_names"`        // 商机名称列表
	CommStyle             string     `json:"comm_style"`               // 沟通风格
	OtherDemands          string     `json:"other_demands"`            // 其他诉求
	IsStop                bool       `json:"is_stop"`                  // 用户回复后是否停止后续节点执行
	IsAIGenerated         bool       `json:"is_ai_generated"`          // 是否由AI生成
	Tasks                 []TaskInfo `json:"tasks"`                    // 任务列表
	FirstTouchAdvanceDays int32      `json:"first_touch_advance_days"` // 第一次触达提前天数
}

type FollowUpConfigJSON struct {
	FollowUpMethods []string              `json:"follow_up_methods"` // 追单方式：coupon-优惠券追单, cart-购物车追单, demand_time-需求时间追单
	CouponConfig    *CouponFollowUpConfig `json:"coupon_config"`     // 优惠券追单配置
	CartConfig      *CartFollowUpConfig   `json:"cart_config"`       // 购物车追单配置
	DemandConfig    *DemandFollowUpConfig `json:"demand_config"`     // 需求时间追单配置
}
type CouponFollowUpConfig struct {
	CouponID    uint64     `json:"coupon_id"`    // 优惠券ID
	CouponName  string     `json:"coupon_name"`  // 优惠券名称
	TriggerType int32      `json:"trigger_type"` // 触发类型
	TriggerTime TimeConfig `json:"trigger_time"` // 触发时间配置
	TaskInfo    TaskInfo   `json:"task_info"`    // 任务配置
}

// CartFollowUpConfig 计划购物车追单配置
type CartFollowUpConfig struct {
	TriggerTime TimeConfig `json:"trigger_time"` // 加入购物车后的触发时间配置
	TaskInfo    TaskInfo   `json:"task_info"`    // 任务配置
}

// DemandFollowUpConfig 计划需求时间追单配置
type DemandFollowUpConfig struct {
	TriggerTime TimeConfig `json:"trigger_time"` // 需求时间前的触发时间配置
	TaskInfo    TaskInfo   `json:"task_info"`    // 任务配置
}

type TimeConfig struct {
	Value int32 `json:"value"` // 数值
	Unit  int32 `json:"unit"`  // 单位：小时或天
}

type TaskInfo struct {
	ID                uint64        `json:"id"`                  // 任务ID
	AdvanceDays       int           `json:"advance_days"`        // 提前天数
	SendTime          string        `json:"send_time"`           // 发送时间，格式：HH:mm:ss
	ContentList       []ContentItem `json:"content_list"`        // 内容列表
	ContentTemplateID uint64        `json:"content_template_id"` // 内容模板ID（可选，如果设置了则使用模板内容）
}

type ContentItem struct {
	MaterialID uint64 `json:"material_id"` // 素材ID
	Type       int32  `json:"type"`        // 内容类型
	Content    string `json:"content"`     // 内容
}

type PlanGroup struct {
	ID                 uint64 `json:"id,string" gorm:"primaryKey;autoIncrement:false"`
	CreatedAt          int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          int64  `json:"updated_at" gorm:"autoUpdateTime"`
	MerchantID         uint64 `gorm:"column:merchant_id;not null;default:0"` // 商户ID
	Name               string `gorm:"column:name;not null"`
	OpportunityCount   uint   `gorm:"column:opportunity_count;default:0"`
	TaskCount          uint   `gorm:"column:task_count;default:0"`
	ConfigJSON         []byte `gorm:"column:config_json;type:json"`
	FollowUpConfigJSON []byte `gorm:"column:follow_up_config_json;type:json"` // 追单配置JSON
	DeletedAt          int64  `gorm:"column:deleted_at;default:0"`
}

func BenchmarkCopier(b *testing.B) {
	type CommonField struct {
		C bool
	}

	type s1 struct {
		A int
		B string
		CommonField
	}

	type s2 struct {
		A int
		B string
		CommonField
	}

	b.Run("copier", func(b *testing.B) {
		struct1 := s1{
			A:           1,
			B:           "2",
			CommonField: CommonField{C: true},
		}
		for i := 0; i < b.N; i++ {
			var struct2 s2
			err := copier.Copy(&struct2, struct1)
			if err != nil {
				require.NoError(b, err)
			}
			require.Equal(b, struct1.A, struct2.A)
			require.Equal(b, struct1.B, struct2.B)
			require.Equal(b, struct1.C, struct2.C)
		}
	})

	b.Run("SimpleMap", func(b *testing.B) {
		struct1 := s1{
			A:           1,
			B:           "2",
			CommonField: CommonField{C: true},
		}
		for i := 0; i < b.N; i++ {
			var struct2 s2
			err := SimpleMap(&struct2, struct1)
			if err != nil {
				require.NoError(b, err)
			}
			require.Equal(b, struct1.A, struct2.A)
			require.Equal(b, struct1.B, struct2.B)
			require.Equal(b, struct1.C, struct2.C)
		}
	})
}

type a struct {
	A *string
}

func (aa *a) Clone() *a {
	s := strings.Clone(*aa.A)
	return &a{
		A: &s,
	}
}

type b2 struct {
	A *string
}

func TestClone(t *testing.T) {
	s := "123"
	aa := a{
		A: &s,
	}
	var b b2

	err := SimpleMap(&b, aa)
	require.NoError(t, err)
	require.False(t, unsafe.StringData(s) != unsafe.StringData(*b.A))
}

type a2 struct {
	A *a
}

type b3 struct {
	A *a
}

func TestClone2(t *testing.T) {
	s := "123"
	a := a2{
		A: &a{
			A: &s,
		},
	}
	var b b3
	err := SimpleMap(&b, a)
	require.NoError(t, err)
	require.True(t, s == *b.A.A)
	require.False(t, unsafe.StringData(s) != unsafe.StringData(*b.A.A))
}

// src所有字段为空时,复制异常
func TestClone3(t *testing.T) {
	type Dest struct {
		A string
	}
	type Src struct {
		A string
	}
	var dest *Dest
	src := Src{}
	err := SimpleMap(&dest, &src)
	require.NoError(t, err)
	require.NotNil(t, dest)
}
