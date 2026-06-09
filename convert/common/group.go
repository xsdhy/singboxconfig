package common

import (
	"singboxconfig/entity"
	"strings"
)

// FilterOutboundGroupTags 根据节点分组规则过滤候选节点标签。
// Include 和 Exclude 均使用逗号分隔关键字；输出顺序保持候选列表原始顺序，
// 这样 sing-box 与 Surge 两条输出链路可以复用完全一致的分组成员规则。
func FilterOutboundGroupTags(groupRule *entity.NodeGroup, tags []string) []string {
	if groupRule == nil || len(tags) == 0 {
		return []string{}
	}
	includes := splitKeywords(groupRule.Include)
	excludes := splitKeywords(groupRule.Exclude)

	// matchedTags 只记录是否命中，最终仍按 tags 顺序输出，避免 map 遍历带来不稳定配置。
	matchedTags := make(map[string]struct{})
	if len(includes) == 0 {
		for _, tag := range tags {
			matchedTags[tag] = struct{}{}
		}
	} else {
		for _, tag := range tags {
			for _, include := range includes {
				if strings.Contains(tag, include) {
					matchedTags[tag] = struct{}{}
					break
				}
			}
		}
	}

	for tag := range matchedTags {
		for _, exclude := range excludes {
			if strings.Contains(tag, exclude) {
				delete(matchedTags, tag)
				break
			}
		}
	}

	result := make([]string, 0, len(matchedTags))
	for _, tag := range tags {
		if _, ok := matchedTags[tag]; ok {
			result = append(result, tag)
		}
	}
	return result
}

// splitKeywords 清理逗号分隔的关键字列表，忽略空白项。
func splitKeywords(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
