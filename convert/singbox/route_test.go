package singbox

import (
	"singboxconfig/entity"
	"testing"
)

// TestGetRouteSkipsRuleWithUnknownOutbound 验证规则集引用的出站不在出站列表中时跳过该条规则，
// 引用已存在出站（含分组与 direct）的规则正常保留，Final 兜底保持 general 不变。
func TestGetRouteSkipsRuleWithUnknownOutbound(t *testing.T) {
	outbounds := []entity.SingBoxOut{
		{Tag: "proxy-node", Type: string(entity.OutboundProtocolShadowsocks)},
		{Tag: "general", Type: string(entity.NodeGroupTypeSelector)},
		{Tag: "direct", Type: string(entity.OutboundProtocolDirect)},
	}
	ruleSets := []*entity.RuleSet{
		{Tag: "rs-group", Outbound: "general", Sort: 10},
		{Tag: "rs-proxy", Outbound: "proxy-node", Sort: 20},
		{Tag: "rs-direct", Outbound: "direct", Sort: 30},
		{Tag: "rs-ghost", Outbound: "missing-node", Sort: 40},
	}

	route := GetRoute("phone", ruleSets, outbounds)

	if route.Final != "general" {
		t.Fatalf("route.Final = %q, want general", route.Final)
	}

	got := make(map[string]string)
	for _, rule := range route.Rules {
		if len(rule.RuleSet) == 1 {
			got[rule.RuleSet[0]] = rule.Outbound
		}
	}

	for tag, wantOutbound := range map[string]string{
		"rs-group":  "general",
		"rs-proxy":  "proxy-node",
		"rs-direct": "direct",
	} {
		if got[tag] != wantOutbound {
			t.Errorf("rule %q outbound = %q, want %q", tag, got[tag], wantOutbound)
		}
	}
	if _, ok := got["rs-ghost"]; ok {
		t.Errorf("rule rs-ghost should be skipped, but found outbound %q", got["rs-ghost"])
	}
}
