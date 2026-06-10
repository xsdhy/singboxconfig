package singbox

import (
	"singboxconfig/entity"
	"strings"
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

	route := GetRoute("phone", "", "", ruleSets, outbounds)

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

// TestGetRouteLocalRuleSetURLReference 验证配置系统 Host 后，
// local 规则集输出为 type:"remote"、format:"source"，且 URL 指向本服务 open 接口并正确转义。
func TestGetRouteLocalRuleSetURLReference(t *testing.T) {
	outbounds := []entity.SingBoxOut{
		{Tag: "general", Type: string(entity.NodeGroupTypeSelector)},
	}
	ruleSets := []*entity.RuleSet{
		{Tag: "geosite-cn", RuleSetType: string(entity.RuleSetTypeLocal), Outbound: "general", Content: `{"version":1,"rules":[{"domain_suffix":["cn"]}]}`},
	}

	route := GetRoute("iphone 15", "tok en", "https://config.example.com/", ruleSets, outbounds)

	if len(route.RuleSet) != 1 {
		t.Fatalf("rule_set len = %d, want 1", len(route.RuleSet))
	}
	rs := route.RuleSet[0]
	if rs.Type != "remote" || rs.Format != "source" {
		t.Errorf("rule_set type/format = %q/%q, want remote/source", rs.Type, rs.Format)
	}
	wantURL := "https://config.example.com/open/rules/geosite-cn/singbox/iphone%2015?token=tok+en"
	if rs.URL != wantURL {
		t.Errorf("rule_set url = %q\nwant %q", rs.URL, wantURL)
	}
}

// TestGetRouteLocalRuleSetFallbackInline 验证未配置系统 Host 时，local 规则集回退为 inline 内联。
func TestGetRouteLocalRuleSetFallbackInline(t *testing.T) {
	outbounds := []entity.SingBoxOut{
		{Tag: "general", Type: string(entity.NodeGroupTypeSelector)},
	}
	ruleSets := []*entity.RuleSet{
		{Tag: "geosite-cn", RuleSetType: string(entity.RuleSetTypeLocal), Outbound: "general", Content: `{"version":1,"rules":[{"domain_suffix":["cn"]}]}`},
	}

	route := GetRoute("iphone", "tok", "", ruleSets, outbounds)

	if len(route.RuleSet) != 1 {
		t.Fatalf("rule_set len = %d, want 1", len(route.RuleSet))
	}
	if route.RuleSet[0].Type != "inline" {
		t.Errorf("rule_set type = %q, want inline", route.RuleSet[0].Type)
	}
	if strings.TrimSpace(route.RuleSet[0].URL) != "" {
		t.Errorf("inline rule_set should not carry url, got %q", route.RuleSet[0].URL)
	}
}

// TestGetRouteSkipsRuleSetWithUnknownOutbound 验证出站不存在的 local 规则集既不出现在 rule_set 也不出现在 rules，
// 避免输出会被客户端下载却不被引用的远程规则集条目（需求 12）。
func TestGetRouteSkipsRuleSetWithUnknownOutbound(t *testing.T) {
	outbounds := []entity.SingBoxOut{
		{Tag: "general", Type: string(entity.NodeGroupTypeSelector)},
	}
	ruleSets := []*entity.RuleSet{
		{Tag: "ghost", RuleSetType: string(entity.RuleSetTypeLocal), Outbound: "missing", Content: `{"version":1,"rules":[{"domain_suffix":["cn"]}]}`},
	}

	route := GetRoute("iphone", "tok", "https://config.example.com", ruleSets, outbounds)

	for _, rs := range route.RuleSet {
		if rs.Tag == "ghost" {
			t.Errorf("rule_set should not contain ghost (outbound missing)")
		}
	}
}
