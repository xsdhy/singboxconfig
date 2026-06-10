package service

import (
	"singboxconfig/entity"
	"testing"
)

// TestFindVisibleRuleSet 覆盖规则集 open 接口的可见性判定：
// tag 命中且全部可见、tag 命中且设备在白名单、tag 命中但设备不在白名单、tag 未命中。
func TestFindVisibleRuleSet(t *testing.T) {
	ruleSets := []*entity.RuleSet{
		{Tag: "all-visible", AbleDevices: ""},
		{Tag: "scoped", AbleDevices: "iphone,ipad"},
		nil, // 列表中可能混入 nil，需安全跳过
	}

	// 全部可见
	if got := findVisibleRuleSet(ruleSets, "all-visible", "anything"); got == nil || got.Tag != "all-visible" {
		t.Errorf("expected all-visible to be returned for any device")
	}

	// 设备在白名单内
	if got := findVisibleRuleSet(ruleSets, "scoped", "iphone"); got == nil || got.Tag != "scoped" {
		t.Errorf("expected scoped to be visible for iphone")
	}

	// 设备不在白名单内：按不可见处理（返回 nil，调用方 404）
	if got := findVisibleRuleSet(ruleSets, "scoped", "android"); got != nil {
		t.Errorf("expected scoped to be invisible for android, got %v", got)
	}

	// tag 未命中
	if got := findVisibleRuleSet(ruleSets, "missing", "iphone"); got != nil {
		t.Errorf("expected nil for missing tag, got %v", got)
	}
}

// TestParseSoftware 验证软件枚举解析覆盖三种合法值与非法值。
func TestParseSoftware(t *testing.T) {
	for _, name := range []string{"singbox", "surge", "shadowrocket"} {
		if _, ok := entity.ParseSoftware(name); !ok {
			t.Errorf("ParseSoftware(%q) should be ok", name)
		}
	}
	for _, name := range []string{"", "clash", "Singbox", "quantumult"} {
		if _, ok := entity.ParseSoftware(name); ok {
			t.Errorf("ParseSoftware(%q) should be rejected", name)
		}
	}
}
