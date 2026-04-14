package singbox

import (
	"encoding/json"
	"singboxconfig/entity"
	"sort"
	"strings"
)

// 匹配说明
//默认规则使用以下匹配逻辑:
//(domain || domain_suffix || domain_keyword || domain_regex || geosite || geoip || ip_cidr || ip_is_private) &&
//(port || port_range) &&
//(source_geoip || source_ip_cidr || source_ip_is_private) &&
//(source_port || source_port_range) &&
//other fields
//另外，引用的规则集可视为被合并，而不是作为一个单独的规则子项。

func GetRoute(device string, ruleSets []*entity.RuleSet) entity.SingRoute {
	route := entity.SingRoute{
		RuleSet:             baseRuleSets(device, ruleSets),
		Final:               "general",
		FindProcess:         false,
		AutoDetectInterface: true,
	}
	//rules
	rules := make([]entity.SingRouteRule, 0)
	//通用基础规则
	rules = append(rules, baseRules()...)

	// 按照 Sort 字段排序
	sort.Slice(ruleSets, func(i, j int) bool {
		return ruleSets[i].Sort < ruleSets[j].Sort
	})

	for _, ruleSet := range ruleSets {
		if ruleSet.AbleDevices != "" && !strings.Contains(ruleSet.AbleDevices, device) {
			continue
		}

		rules = append(rules, entity.SingRouteRule{
			RuleSet:  []string{ruleSet.Tag},
			Outbound: ruleSet.Outbound,
		})
	}

	// if device == "office" || device == "phone" {
	// 	rules = append(rules, entity.SingRouteRule{
	// 		IPCIDR: []string{
	// 			"192.168.10.0/24",
	// 		},
	// 		WifiSSID: []string{
	// 			"!DoubleTang_5G",
	// 			"!DoubleTang",
	// 		},
	// 		Outbound: "wg-ep",
	// 	})
	// }

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

func baseRuleSets(device string, ruleSets []*entity.RuleSet) []entity.SingRuleSet {
	ruleset := make([]entity.SingRuleSet, 0)

	for _, ruleSet := range ruleSets {
		if ruleSet.AbleDevices != "" && !strings.Contains(ruleSet.AbleDevices, device) {
			continue
		}

		if ruleSet.RuleSetType == "remote" {
			ruleset = append(ruleset, entity.SingRuleSet{
				Type:           "remote",
				Tag:            ruleSet.Tag,
				Format:         ruleSet.Format,
				URL:            ruleSet.URL,
				DownloadDetour: ruleSet.DownloadDetour,
			})
		} else {
			var rules json.RawMessage
			if err := json.Unmarshal([]byte(ruleSet.Content), &rules); err != nil {
				continue
			}
			ruleset = append(ruleset, entity.SingRuleSet{
				Type:  "inline",
				Tag:   ruleSet.Tag,
				Rules: rules,
			})
		}
	}

	return ruleset
}
