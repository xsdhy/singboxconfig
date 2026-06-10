package singbox

import (
	"singboxconfig/convert/common"
	"singboxconfig/entity"
)

// GetOutbounds 基于已经准备好的缓存出站构建最终 sing-box outbounds。
// 这里不再承担订阅拉取职责，只负责分组拼装和固定出站追加。
// deviceCode 用于按设备解析分组类型（selector / urltest），为空时只使用分组默认类型。
func GetOutbounds(outbounds []entity.SingBoxOut, groupRules []*entity.NodeGroup, deviceCode string) []entity.SingBoxOut {
	result := append([]entity.SingBoxOut(nil), outbounds...)

	groupOutbounds := constructOutboundGroup(groupRules, getTagsFromOutbounds(result), deviceCode)
	result = append(result, groupOutbounds...)
	result = append(result, entity.SingBoxOut{Type: "direct", Tag: "direct"})
	return result
}

func constructOutboundGroup(groupRules []*entity.NodeGroup, tags []string, deviceCode string) []entity.SingBoxOut {
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

		if len(filteredTags) == 0 {
			continue
		}

		// 按设备解析最终分组类型，命中设备覆盖时使用覆盖类型，否则回退默认 GroupType。
		// 返回值已规范化为 selector / urltest，select 兼容值会被归一化为 selector。
		groupType := common.ResolveGroupType(groupRule, deviceCode)

		// 创建出站配置
		item := entity.SingBoxOut{
			Tag:       groupRule.Tag,
			Type:      string(groupType),
			Outbounds: filteredTags,
		}

		// 根据类型设置特定配置
		switch groupType {
		case entity.NodeGroupTypeURLTest:
			// 使用局部变量处理默认探测地址，避免回写修改原始 groupRule.TestURL。
			testURL := groupRule.TestURL
			if testURL == "" {
				testURL = "https://www.gstatic.com/generate_204"
			}
			item.URL = testURL
			item.Interval = "10m"
			item.Tolerance = 50 // 添加容差值
		case entity.NodeGroupTypeSelector:
			item.Default = filteredTags[0] // 设置默认节点
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
