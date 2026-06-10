package common

import (
	"reflect"
	"singboxconfig/entity"
	"testing"
)

// TestParseDeviceTypeOverrides 覆盖设备级类型覆盖规则字符串的各种解析分支。
func TestParseDeviceTypeOverrides(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want map[string]entity.NodeGroupType
	}{
		{
			name: "空字符串返回空映射",
			raw:  "",
			want: map[string]entity.NodeGroupType{},
		},
		{
			name: "多条规则正常解析",
			raw:  "phone:selector,gateway:urltest",
			want: map[string]entity.NodeGroupType{
				"phone":   entity.NodeGroupTypeSelector,
				"gateway": entity.NodeGroupTypeURLTest,
			},
		},
		{
			name: "两端空白被去除",
			raw:  " phone : selector , gateway : urltest ",
			want: map[string]entity.NodeGroupType{
				"phone":   entity.NodeGroupTypeSelector,
				"gateway": entity.NodeGroupTypeURLTest,
			},
		},
		{
			name: "select 兼容值归一化为 selector",
			raw:  "phone:select",
			want: map[string]entity.NodeGroupType{
				"phone": entity.NodeGroupTypeSelector,
			},
		},
		{
			name: "非法类型被忽略",
			raw:  "phone:bad,gateway:urltest",
			want: map[string]entity.NodeGroupType{
				"gateway": entity.NodeGroupTypeURLTest,
			},
		},
		{
			name: "缺少冒号或空设备编码被忽略",
			raw:  "phone,:urltest,gateway:urltest",
			want: map[string]entity.NodeGroupType{
				"gateway": entity.NodeGroupTypeURLTest,
			},
		},
		{
			name: "目标类型为空被忽略",
			raw:  "phone:,gateway:selector",
			want: map[string]entity.NodeGroupType{
				"gateway": entity.NodeGroupTypeSelector,
			},
		},
		{
			name: "重复设备编码后者覆盖前者",
			raw:  "phone:selector,phone:urltest",
			want: map[string]entity.NodeGroupType{
				"phone": entity.NodeGroupTypeURLTest,
			},
		},
		{
			name: "设备编码大小写敏感",
			raw:  "Phone:selector,phone:urltest",
			want: map[string]entity.NodeGroupType{
				"Phone": entity.NodeGroupTypeSelector,
				"phone": entity.NodeGroupTypeURLTest,
			},
		},
		{
			name: "目标类型只取第一个冒号后的内容",
			raw:  "phone:selector:extra",
			want: map[string]entity.NodeGroupType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDeviceTypeOverrides(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDeviceTypeOverrides(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

// TestResolveGroupType 覆盖按设备解析分组类型的命中、回退与规范化逻辑。
func TestResolveGroupType(t *testing.T) {
	tests := []struct {
		name       string
		group      *entity.NodeGroup
		deviceCode string
		want       entity.NodeGroupType
	}{
		{
			name:       "group 为 nil 回退 selector",
			group:      nil,
			deviceCode: "phone",
			want:       entity.NodeGroupTypeSelector,
		},
		{
			name:       "命中设备覆盖为 urltest",
			group:      &entity.NodeGroup{GroupType: "selector", DeviceTypeOverrides: "gateway:urltest"},
			deviceCode: "gateway",
			want:       entity.NodeGroupTypeURLTest,
		},
		{
			name:       "未命中覆盖回退默认 urltest",
			group:      &entity.NodeGroup{GroupType: "urltest", DeviceTypeOverrides: "gateway:selector"},
			deviceCode: "phone",
			want:       entity.NodeGroupTypeURLTest,
		},
		{
			name:       "覆盖为 select 规范化为 selector",
			group:      &entity.NodeGroup{GroupType: "urltest", DeviceTypeOverrides: "phone:select"},
			deviceCode: "phone",
			want:       entity.NodeGroupTypeSelector,
		},
		{
			name:       "默认类型 select 规范化为 selector",
			group:      &entity.NodeGroup{GroupType: "select"},
			deviceCode: "phone",
			want:       entity.NodeGroupTypeSelector,
		},
		{
			name:       "默认类型非法回退 selector",
			group:      &entity.NodeGroup{GroupType: "bad"},
			deviceCode: "phone",
			want:       entity.NodeGroupTypeSelector,
		},
		{
			name:       "设备编码为空只看默认类型",
			group:      &entity.NodeGroup{GroupType: "urltest", DeviceTypeOverrides: "phone:selector"},
			deviceCode: "",
			want:       entity.NodeGroupTypeURLTest,
		},
		{
			name:       "覆盖中的非法类型被忽略后回退默认",
			group:      &entity.NodeGroup{GroupType: "urltest", DeviceTypeOverrides: "phone:bad"},
			deviceCode: "phone",
			want:       entity.NodeGroupTypeURLTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveGroupType(tt.group, tt.deviceCode)
			if got != tt.want {
				t.Errorf("ResolveGroupType() = %v, want %v", got, tt.want)
			}
		})
	}
}
