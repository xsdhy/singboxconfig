package singbox

import (
	"encoding/json"
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
func GetRoute(device string, ruleSets []*entity.RuleSet, outbounds []entity.SingBoxOut) entity.SingRoute {
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

	// 已存在的出站标签集合，用于规则引用校验；FINAL 兜底与基础规则不参与校验。
	outboundTags := make(map[string]struct{}, len(outbounds))
	for _, outbound := range outbounds {
		if outbound.Tag != "" {
			outboundTags[outbound.Tag] = struct{}{}
		}
	}

	// 按照 Sort 字段排序
	sort.Slice(ruleSets, func(i, j int) bool {
		return ruleSets[i].Sort < ruleSets[j].Sort
	})

	for _, ruleSet := range ruleSets {
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
