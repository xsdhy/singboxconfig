package singbox

import (
	"singboxconfig/entity"
	"strings"
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
		case "urltest":
			if groupRule.TestURL == "" {
				groupRule.TestURL = "https://www.gstatic.com/generate_204"
			}
			item.URL = groupRule.TestURL
			item.Interval = "10m"
			item.Tolerance = 50 // 添加容差值
		case "selector":
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
	if groupRule == nil || len(tags) == 0 {
		return []string{}
	}
	includes := make([]string, 0)
	excludes := make([]string, 0)
	if groupRule.Include != "" {
		if strings.Contains(groupRule.Include, ",") {
			includes = strings.Split(groupRule.Include, ",")
		} else {
			includes = append(includes, groupRule.Include)
		}
	}
	if groupRule.Exclude != "" {
		if strings.Contains(groupRule.Exclude, ",") {
			excludes = strings.Split(groupRule.Exclude, ",")
		} else {
			excludes = append(excludes, groupRule.Exclude)
		}
	}

	// 使用map来存储匹配的标签，避免重复
	matchedTags := make(map[string]struct{})

	// 如果没有包含规则，默认包含所有标签
	if len(includes) == 0 {
		for _, tag := range tags {
			matchedTags[tag] = struct{}{}
		}
	} else {
		// 根据包含规则筛选标签
		for _, tag := range tags {
			for _, include := range includes {
				if strings.Contains(tag, include) {
					matchedTags[tag] = struct{}{}
					break
				}
			}
		}
	}

	// 如果有排除规则，从匹配的标签中移除
	if len(excludes) > 0 {
		for tag := range matchedTags {
			for _, exclude := range excludes {
				if strings.Contains(tag, exclude) {
					delete(matchedTags, tag)
					break
				}
			}
		}
	}

	// 按原始 tags 顺序输出，避免 map 遍历导致生成结果和测试顺序不稳定。
	result := make([]string, 0, len(matchedTags))
	for _, tag := range tags {
		if _, ok := matchedTags[tag]; ok {
			result = append(result, tag)
		}
	}

	return result
}

func getTagsFromOutbounds(outbounds []entity.SingBoxOut) []string {
	tags := make([]string, 0)
	for _, outbound := range outbounds {
		tags = append(tags, outbound.Tag)
	}
	return tags
}
