package singbox

import (
	"reflect"
	"singboxconfig/entity"
	"testing"
)

func TestOutboundGroup(t *testing.T) {
	tests := []struct {
		name       string
		groupRules []*entity.NodeGroup
		tags       []string
		deviceCode string
		want       []entity.SingBoxOut
	}{
		{
			name: "基本功能测试",
			groupRules: []*entity.NodeGroup{
				{
					Tag:       "general",
					GroupType: "urltest",
					Include:   "香港",
				},
				{
					Tag:       "developer",
					GroupType: "urltest",
					Include:   "新加坡",
				},
			},
			tags: []string{"香港01", "香港02", "新加坡01", "新加坡02", "美国01"},
			want: []entity.SingBoxOut{
				{
					Tag:       "general",
					Type:      "urltest",
					Outbounds: []string{"香港01", "香港02"},
					URL:       "https://www.gstatic.com/generate_204",
					Interval:  "10m",
					Tolerance: 50,
				},
				{
					Tag:       "developer",
					Type:      "urltest",
					Outbounds: []string{"新加坡01", "新加坡02"},
					URL:       "https://www.gstatic.com/generate_204",
					Interval:  "10m",
					Tolerance: 50,
				},
			},
		},
		{
			name:       "空输入测试",
			groupRules: []*entity.NodeGroup{},
			tags:       []string{},
			want:       []entity.SingBoxOut{},
		},
		{
			name: "无匹配节点测试",
			groupRules: []*entity.NodeGroup{
				{
					Tag:       "general",
					GroupType: "urltest",
					Include:   "香港",
				},
			},
			tags: []string{"美国01", "美国02"},
			want: []entity.SingBoxOut{},
		},
		{
			name: "设备覆盖为 urltest",
			groupRules: []*entity.NodeGroup{
				{
					Tag:                 "general",
					GroupType:           "selector",
					Include:             "香港",
					DeviceTypeOverrides: "gateway:urltest",
				},
			},
			tags:       []string{"香港01", "香港02"},
			deviceCode: "gateway",
			want: []entity.SingBoxOut{
				{
					Tag:       "general",
					Type:      "urltest",
					Outbounds: []string{"香港01", "香港02"},
					URL:       "https://www.gstatic.com/generate_204",
					Interval:  "10m",
					Tolerance: 50,
				},
			},
		},
		{
			name: "未命中设备回退默认 selector",
			groupRules: []*entity.NodeGroup{
				{
					Tag:                 "general",
					GroupType:           "selector",
					Include:             "香港",
					DeviceTypeOverrides: "gateway:urltest",
				},
			},
			tags:       []string{"香港01", "香港02"},
			deviceCode: "phone",
			want: []entity.SingBoxOut{
				{
					Tag:       "general",
					Type:      "selector",
					Outbounds: []string{"香港01", "香港02"},
					Default:   "香港01",
				},
			},
		},
		{
			name: "默认 select 规范化为 selector",
			groupRules: []*entity.NodeGroup{
				{
					Tag:       "general",
					GroupType: "select",
					Include:   "香港",
				},
			},
			tags:       []string{"香港01"},
			deviceCode: "phone",
			want: []entity.SingBoxOut{
				{
					Tag:       "general",
					Type:      "selector",
					Outbounds: []string{"香港01"},
					Default:   "香港01",
				},
			},
		},
		{
			name: "设备覆盖为 urltest 且 testURL 为空使用默认地址且不回写原始值",
			groupRules: []*entity.NodeGroup{
				{
					Tag:                 "general",
					GroupType:           "selector",
					Include:             "香港",
					DeviceTypeOverrides: "gateway:urltest",
				},
			},
			tags:       []string{"香港01"},
			deviceCode: "gateway",
			want: []entity.SingBoxOut{
				{
					Tag:       "general",
					Type:      "urltest",
					Outbounds: []string{"香港01"},
					URL:       "https://www.gstatic.com/generate_204",
					Interval:  "10m",
					Tolerance: 50,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := constructOutboundGroup(tt.groupRules, tt.tags, tt.deviceCode)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OutboundGroup() = %v, want %v", got, tt.want)
			}
			// 校验转换过程不会回写修改原始 NodeGroup.TestURL。
			for _, rule := range tt.groupRules {
				if rule.TestURL != "" {
					t.Errorf("constructOutboundGroup 不应修改原始 TestURL，got=%q", rule.TestURL)
				}
			}
		})
	}
}

func TestOutboundRuleFilter(t *testing.T) {
	tests := []struct {
		name      string
		groupRule *entity.NodeGroup
		tags      []string
		want      []string
	}{
		{
			name: "基本包含规则测试",
			groupRule: &entity.NodeGroup{
				Include: "香港",
			},
			tags: []string{"香港01", "香港02", "新加坡01", "新加坡02"},
			want: []string{"香港01", "香港02"},
		},
		{
			name: "基本排除规则测试",
			groupRule: &entity.NodeGroup{
				Exclude: "官网",
			},
			tags: []string{"香港01", "香港02", "官网节点", "新加坡01"},
			want: []string{"香港01", "香港02", "新加坡01"},
		},
		{
			name: "包含和排除规则组合测试",
			groupRule: &entity.NodeGroup{
				Include: "香港",
				Exclude: "官网",
			},
			tags: []string{"香港01", "香港官网", "新加坡01"},
			want: []string{"香港01"},
		},
		{
			name:      "空规则测试",
			groupRule: &entity.NodeGroup{},
			tags:      []string{"香港01", "新加坡01"},
			want:      []string{"香港01", "新加坡01"},
		},
		{
			name: "空标签测试",
			groupRule: &entity.NodeGroup{
				Include: "香港",
			},
			tags: []string{},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := outboundGroupRuleFilter(tt.groupRule, tt.tags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("outboundRuleFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
