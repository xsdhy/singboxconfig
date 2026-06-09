package singbox

import (
	"singboxconfig/convert/common"
	"singboxconfig/entity"
)

// GetOutbounds 基于已经准备好的缓存出站构建最终 sing-box outbounds。
// 这里不再承担订阅拉取职责，只负责分组拼装和固定出站追加。
func GetOutbounds(outbounds []entity.SingBoxOut, groupRules []*entity.NodeGroup) []entity.SingBoxOut {
	result := append([]entity.SingBoxOut(nil), outbounds...)

	groupOutbounds := constructOutboundGroup(groupRules, getTagsFromOutbounds(result))
	result = append(result, groupOutbounds...)
	result = append(result, entity.SingBoxOut{Type: "direct", Tag: "direct"})
	return result
}

func constructOutboundGroup(groupRules []*entity.NodeGroup, tags []string) []entity.SingBoxOut {
	if len(groupRules) == 0 {
		return []entity.SingBoxOut{}
	}
	groupOutbounds := make([]entity.SingBoxOut, 0, len(groupRules))
	for _, groupRule := range groupRules {
		// 验证必要的字段
		if groupRule.Tag == "" {
			continue
		}
		// 获取过滤后的标签
		filteredTags := outboundGroupRuleFilter(groupRule, tags)

		// 创建出站配置
		item := entity.SingBoxOut{
			Tag:       groupRule.Tag,
			Type:      groupRule.GroupType,
			Outbounds: filteredTags,
		}

		if len(filteredTags) == 0 {
			continue
		}

		// 根据类型设置特定配置
		switch groupRule.GroupType {
		case string(entity.NodeGroupTypeURLTest):
			if groupRule.TestURL == "" {
				groupRule.TestURL = "https://www.gstatic.com/generate_204"
			}
			item.URL = groupRule.TestURL
			item.Interval = "10m"
			item.Tolerance = 50 // 添加容差值
		case string(entity.NodeGroupTypeSelector), string(entity.NodeGroupTypeSelect):
			if len(filteredTags) > 0 {
				item.Default = filteredTags[0] // 设置默认节点
			}
		}

		groupOutbounds = append(groupOutbounds, item)
	}
	return groupOutbounds
}

// outboundGroupRuleFilter 根据规则过滤标签
func outboundGroupRuleFilter(groupRule *entity.NodeGroup, tags []string) []string {
	return common.FilterOutboundGroupTags(groupRule, tags)
}

func getTagsFromOutbounds(outbounds []entity.SingBoxOut) []string {
	tags := make([]string, 0)
	for _, outbound := range outbounds {
		tags = append(tags, outbound.Tag)
	}
	return tags
}
