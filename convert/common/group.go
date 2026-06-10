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

// ParseDeviceTypeOverrides 解析节点分组的设备级类型覆盖规则字符串。
// 入参格式为逗号分隔的 "设备编码:分组类型" 列表，例如 "phone:selector,gateway:urltest"。
// 返回 map[设备编码]规范化后的分组类型，解析规则遵循需求约定：
//   - 以英文逗号 "," 分隔多条覆盖规则；
//   - 每条规则以第一个英文冒号 ":" 分隔设备编码与目标类型；
//   - 设备编码与目标类型两端空白会被 TrimSpace 去除；
//   - 设备编码为空、目标类型为空、缺少冒号的规则直接忽略；
//   - 目标类型只接受 selector / select / urltest，非法值忽略；
//   - select 会被规范化为 selector，便于下游统一判断；
//   - 同一设备编码多次出现时，后出现的规则覆盖先出现的规则；
//   - 设备编码大小写敏感，不校验设备是否真实存在。
func ParseDeviceTypeOverrides(raw string) map[string]entity.NodeGroupType {
	overrides := make(map[string]entity.NodeGroupType)
	if strings.TrimSpace(raw) == "" {
		return overrides
	}
	for _, rule := range strings.Split(raw, ",") {
		// 缺少冒号的规则直接忽略。
		deviceCode, rawType, ok := strings.Cut(rule, ":")
		if !ok {
			continue
		}
		deviceCode = strings.TrimSpace(deviceCode)
		groupType, valid := normalizeGroupType(rawType)
		// 设备编码为空或目标类型非法时忽略该条规则。
		if deviceCode == "" || !valid {
			continue
		}
		// 后出现的规则直接覆盖先出现的规则。
		overrides[deviceCode] = groupType
	}
	return overrides
}

// ResolveGroupType 根据设备编码解析节点分组最终应使用的分组类型。
// 命中设备覆盖规则时返回覆盖类型，否则回退到分组默认的 GroupType。
// 返回值已规范化为 NodeGroupTypeSelector 或 NodeGroupTypeURLTest 两种规范类型，
// 兼容值 select 会被归一化为 selector，供 sing-box / Surge / Shadowrocket 三条输出链路统一使用。
func ResolveGroupType(group *entity.NodeGroup, deviceCode string) entity.NodeGroupType {
	if group == nil {
		return entity.NodeGroupTypeSelector
	}
	// 优先尝试命中设备级覆盖规则。
	if deviceCode != "" {
		overrides := ParseDeviceTypeOverrides(group.DeviceTypeOverrides)
		if override, ok := overrides[deviceCode]; ok {
			return override
		}
	}
	// 未命中覆盖时回退到默认 GroupType，并做规范化。
	if groupType, valid := normalizeGroupType(group.GroupType); valid {
		return groupType
	}
	// 默认类型为空或非法时，回退为 selector，与现有“非法值回退默认”的实现风格一致。
	return entity.NodeGroupTypeSelector
}

// normalizeGroupType 校验并规范化分组类型字符串。
// 仅接受 selector / select / urltest 三种合法值；
// select 归一化为 selector；非法或空值返回 valid=false。
func normalizeGroupType(raw string) (entity.NodeGroupType, bool) {
	switch entity.NodeGroupType(strings.TrimSpace(raw)) {
	case entity.NodeGroupTypeSelector, entity.NodeGroupTypeSelect:
		return entity.NodeGroupTypeSelector, true
	case entity.NodeGroupTypeURLTest:
		return entity.NodeGroupTypeURLTest, true
	default:
		return "", false
	}
}
