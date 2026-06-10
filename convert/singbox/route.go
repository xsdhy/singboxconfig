package singbox

import (
	"encoding/json"
	"singboxconfig/convert/ruleset"
	"singboxconfig/entity"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// 匹配说明
//默认规则使用以下匹配逻辑:
//(domain || domain_suffix || domain_keyword || domain_regex || geosite || geoip || ip_cidr || ip_is_private) &&
//(port || port_range) &&
//(source_geoip || source_ip_cidr || source_ip_is_private) &&
//(source_port || source_port_range) &&
//other fields
//另外，引用的规则集可视为被合并，而不是作为一个单独的规则子项。

// GetRoute 生成 sing-box 路由配置。
// outbounds 是当前设备最终的出站列表（含分组出站与 direct），用于校验规则引用的出站是否真实存在；
// 规则引用了不存在的出站时直接跳过该条，避免生成指向空出站的路由规则。
//
// deviceToken 与 systemHost 用于把 local / inline 规则集改为指向本服务规则集 open 接口的远程 URL 引用：
// 仅当 systemHost 非空且规则集内容可解析时才输出远程引用，否则回退到原 inline 内联行为，保证旧部署零配置可用。
//
// 为避免输出“未被任何有效路由引用、但仍会被客户端下载”的规则集，这里先算出当前设备真正有效的规则集
// （AbleDevices 通过 + Outbound 存在），再对同一批有效规则集同时输出 route.rule_set 与 route.rules。
func GetRoute(device string, deviceToken string, systemHost string, ruleSets []*entity.RuleSet, outbounds []entity.SingBoxOut) entity.SingRoute {
	// 已存在的出站标签集合，用于规则引用校验；FINAL 兜底与基础规则不参与校验。
	outboundTags := make(map[string]struct{}, len(outbounds))
	for _, outbound := range outbounds {
		if outbound.Tag != "" {
			outboundTags[outbound.Tag] = struct{}{}
		}
	}

	// 按照 Sort 字段排序后，过滤出当前设备真正有效的规则集。
	sorted := append([]*entity.RuleSet(nil), ruleSets...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Sort < sorted[j].Sort
	})

	effective := make([]*entity.RuleSet, 0, len(sorted))
	for _, ruleSet := range sorted {
		if ruleSet.AbleDevices != "" && !strings.Contains(ruleSet.AbleDevices, device) {
			continue
		}
		// 规则指定了出站，但该出站在当前设备的出站列表中不存在时跳过该条规则。
		if ruleSet.Outbound != "" {
			if _, ok := outboundTags[ruleSet.Outbound]; !ok {
				logrus.Warnf("GetRoute: ruleset outbound not found in outbounds, skip rule: tag=%s outbound=%s", ruleSet.Tag, ruleSet.Outbound)
				continue
			}
		}
		effective = append(effective, ruleSet)
	}

	// 生成 rule_set 条目。effective 已按出站存在性过滤，因此不会输出“出站不存在、不会被 route.rules 引用”
	// 却仍会被客户端下载的远程规则集条目（需求 12）。
	ruleSetEntries := baseRuleSets(device, deviceToken, systemHost, effective)

	route := entity.SingRoute{
		RuleSet:             ruleSetEntries,
		Final:               "general",
		FindProcess:         false,
		AutoDetectInterface: true,
	}

	//rules
	rules := make([]entity.SingRouteRule, 0)
	//通用基础规则
	rules = append(rules, baseRules()...)

	for _, ruleSet := range effective {
		rules = append(rules, entity.SingRouteRule{
			RuleSet:  []string{ruleSet.Tag},
			Outbound: ruleSet.Outbound,
		})
	}

	route.Rules = rules

	return route
}

func baseRules() []entity.SingRouteRule {
	return []entity.SingRouteRule{
		{
			Protocol: "dns",
			Action:   "hijack-dns",
		},
		{
			ClashMode: "direct",
			Outbound:  "direct",
		},
		{
			ClashMode: "global",
			Outbound:  "select",
		},
		{
			Protocol: "quic",
			Action:   "reject",
		},
		// {
		// 	Protocol: "udp",
		// 	Outbound: "direct",
		// },
		// {
		// 	Port:     []int{22}, //主要是应对github推送代码
		// 	Outbound: "direct",
		// },
	}
}

// baseRuleSets 把已过滤的有效规则集渲染为 sing-box route.rule_set 条目。
// 入参 ruleSets 已按 AbleDevices 与出站存在性过滤、排序，本函数只负责 remote / local 两种形态的输出：
//   - remote：直接引用 RuleSet.URL；
//   - local / inline：systemHost 非空且内容可规范化时，输出 type:"remote" + 指向本服务 open 接口的 url；
//     否则回退到 type:"inline" 内联（内容非法则跳过该 rule_set 条目）。
func baseRuleSets(device string, deviceToken string, systemHost string, ruleSets []*entity.RuleSet) []entity.SingRuleSet {
	entries := make([]entity.SingRuleSet, 0, len(ruleSets))

	for _, ruleSet := range ruleSets {
		if entity.RuleSetType(ruleSet.RuleSetType) == entity.RuleSetTypeRemote {
			entries = append(entries, entity.SingRuleSet{
				Type:           "remote",
				Tag:            ruleSet.Tag,
				Format:         ruleSet.Format,
				URL:            ruleSet.URL,
				DownloadDetour: ruleSet.DownloadDetour,
			})
			continue
		}

		// local / inline：系统 Host 可用且内容可规范化时，改为指向本服务 open 接口的远程 URL 引用。
		if strings.TrimSpace(systemHost) != "" {
			if _, err := ruleset.NormalizeSingbox(ruleSet.Content); err == nil {
				entries = append(entries, entity.SingRuleSet{
					Type:           "remote",
					Tag:            ruleSet.Tag,
					Format:         "source",
					URL:            ruleset.BuildRuleSetURL(systemHost, ruleSet.Tag, entity.SoftwareSingbox, device, deviceToken),
					DownloadDetour: ruleSet.DownloadDetour,
				})
				continue
			}
			// 内容非法时不生成指向坏内容的远程 URL，回退到下面的 inline 内联逻辑（同样会跳过坏内容）。
			logrus.Warnf("baseRuleSets: local ruleset content invalid, fallback to inline: tag=%s", ruleSet.Tag)
		}

		// 回退：内联展开。内容非法时跳过该 rule_set 条目。
		var rules json.RawMessage
		if err := json.Unmarshal([]byte(ruleSet.Content), &rules); err != nil {
			continue
		}
		entries = append(entries, entity.SingRuleSet{
			Type:  "inline",
			Tag:   ruleSet.Tag,
			Rules: rules,
		})
	}

	return entries
}
