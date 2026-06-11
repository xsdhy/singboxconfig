// Package ruleset 统一承载规则集（entity.RuleSet.Content）的解析、规范化与各客户端渲染能力。
//
// 之前 convert/surge 与 convert/shadowrocket 各自复制了一份“把 sing-box source 规则展开成
// 逐行规则”的逻辑，本包把这些能力集中起来，供三处复用：
//   - 规则集 open 接口（service.GetRulesBySoftware）：按软件输出单个规则集文件内容；
//   - 生成链路：当系统 Host 可用时，把 local / inline 规则集改为指向本服务的远程 URL 引用。
//
// 本包是纯函数实现，不访问数据库、不发起网络请求，因此解析、规范化与渲染均可直接单元测试。
package ruleset

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"singboxconfig/entity"
	"strings"
)

// ErrInvalidContent 表示规则集 Content 整体不是合法的 JSON 对象或数组。
// 调用方在 open 接口场景下据此返回 500；在生成链路场景下据此回退到原展开逻辑或跳过。
var ErrInvalidContent = errors.New("ruleset: content is not a valid json object or array")

// LineRuleType 表示 Surge / Shadowrocket 规则集文件里一行规则的类型列。
// 两个客户端的规则类型关键字一致，因此共用同一套枚举，避免裸字符串散落各处。
type LineRuleType string

const (
	// LineRuleDomain 表示域名精确匹配规则，对应 sing-box source 的 domain。
	LineRuleDomain LineRuleType = "DOMAIN"
	// LineRuleDomainSuffix 表示域名后缀匹配规则，对应 domain_suffix。
	LineRuleDomainSuffix LineRuleType = "DOMAIN-SUFFIX"
	// LineRuleDomainKeyword 表示域名关键字匹配规则，对应 domain_keyword。
	LineRuleDomainKeyword LineRuleType = "DOMAIN-KEYWORD"
	// LineRuleDomainRegex 表示域名正则匹配规则，对应 domain_regex。
	LineRuleDomainRegex LineRuleType = "DOMAIN-REGEX"
	// LineRuleIPCIDR 表示 IPv4 CIDR 匹配规则，对应 ip_cidr 中的 IPv4 项。
	LineRuleIPCIDR LineRuleType = "IP-CIDR"
	// LineRuleIPCIDR6 表示 IPv6 CIDR 匹配规则，对应 ip_cidr 中的 IPv6 项。
	LineRuleIPCIDR6 LineRuleType = "IP-CIDR6"
	// LineRuleGEOIP 表示国家或地区 IP 库匹配规则，对应 geoip。
	LineRuleGEOIP LineRuleType = "GEOIP"
	// LineRuleProcessName 表示进程匹配规则，对应 sing-box source 的 process_name / process_path。
	// Surge / Shadowrocket 的 PROCESS-NAME 既接受进程名也接受可执行文件全路径，故两字段共用同一类型。
	LineRuleProcessName LineRuleType = "PROCESS-NAME"
)

// geoipNoResolve 是 GEOIP 规则行附带的 no-resolve 选项，提示客户端不要为匹配而触发 DNS 解析。
const geoipNoResolve = "no-resolve"

// InlineThreshold 是 local / inline 规则集改用远程 URL 引用所需的最小规则条数。
// 规则条数小于该值时，生成链路直接把规则内联展开，避免为极少量规则额外引入一次远程请求。
const InlineThreshold = 3

// 以下常量是 sing-box source 规则集里本包第一版支持转换的字段名。
// 统一在这里声明，既用于按固定顺序渲染，也用于识别“未支持字段”从而记录 warning。
const (
	fieldDomain        = "domain"
	fieldDomainSuffix  = "domain_suffix"
	fieldDomainKeyword = "domain_keyword"
	fieldDomainRegex   = "domain_regex"
	fieldIPCIDR        = "ip_cidr"
	fieldGEOIP         = "geoip"
	fieldProcessName   = "process_name"
	fieldProcessPath   = "process_path"
)

// supportedFields 是已支持字段的集合，用于在渲染逐行规则时判定未知字段并记录 warning。
var supportedFields = map[string]struct{}{
	fieldDomain:        {},
	fieldDomainSuffix:  {},
	fieldDomainKeyword: {},
	fieldDomainRegex:   {},
	fieldIPCIDR:        {},
	fieldGEOIP:         {},
	fieldProcessName:   {},
	fieldProcessPath:   {},
}

// singboxSourceRuleSet 是 sing-box source 规则集文件的规范化形态：{"version":1,"rules":[...]}。
type singboxSourceRuleSet struct {
	// Version 固定为 1，对应 sing-box source 规则集版本号。
	Version int `json:"version"`
	// Rules 保留原始规则对象，规范化时不丢弃任何字段。
	Rules []json.RawMessage `json:"rules"`
}

// parseRawRules 把 RuleSet.Content 解析为规则对象数组，兼容三种输入形态：
//  1. 完整 sing-box source 规则集：{"version":1,"rules":[...]}；
//  2. 裸 rules 数组：[{...}, {...}]；
//  3. 单条规则对象：{"domain_suffix":["example.com"]}。
//
// 返回的每个元素都是原始 JSON 规则对象（json.RawMessage），便于规范化输出时无损保留全部字段。
func parseRawRules(content string) ([]json.RawMessage, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, ErrInvalidContent
	}

	switch trimmed[0] {
	case '[':
		// 形态 2：裸 rules 数组。
		var arr []json.RawMessage
		if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
			return nil, err
		}
		return arr, nil
	case '{':
		// 形态 1 或 3：先判断是否带 rules 键。
		var obj map[string]json.RawMessage
		if err := json.Unmarshal([]byte(trimmed), &obj); err != nil {
			return nil, err
		}
		if rulesRaw, ok := obj["rules"]; ok {
			var rules []json.RawMessage
			if err := json.Unmarshal(rulesRaw, &rules); err != nil {
				return nil, err
			}
			return rules, nil
		}
		// 不带 rules 键时，把整个对象视为单条规则。
		return []json.RawMessage{json.RawMessage(trimmed)}, nil
	default:
		// 顶层既不是对象也不是数组，视为非法内容。
		return nil, ErrInvalidContent
	}
}

// NormalizeSingbox 把 RuleSet.Content 规范化为 sing-box source 规则集 JSON：{"version":1,"rules":[...]}。
// Content 非法时返回 ErrInvalidContent（或底层 JSON 错误），由调用方决定返回 500 还是回退。
func NormalizeSingbox(content string) ([]byte, error) {
	rules, err := parseRawRules(content)
	if err != nil {
		return nil, err
	}
	if rules == nil {
		rules = []json.RawMessage{}
	}
	return json.Marshal(singboxSourceRuleSet{Version: 1, Rules: rules})
}

// InlineRules 解析 RuleSet.Content 并返回其规则对象数组，供 sing-box inline 规则集（route.rule_set 的 rules 字段）直接内联。
// 兼容完整 source 规则集（{"version":1,"rules":[...]}）、裸 rules 数组与单条规则对象三种形态，
// 始终返回剥离外层包装后的 rules 数组；Content 非法时返回 ErrInvalidContent（或底层 JSON 错误）。
func InlineRules(content string) ([]json.RawMessage, error) {
	rules, err := parseRawRules(content)
	if err != nil {
		return nil, err
	}
	if rules == nil {
		rules = []json.RawMessage{}
	}
	return rules, nil
}

// RenderLines 把 RuleSet.Content 渲染为 Surge / Shadowrocket 规则集文件的逐行内容，
// 每行形如 `类型,值`（不含出站策略列）。两个客户端规则类型关键字一致，故共用同一渲染逻辑。
//
// 返回值：
//   - lines：渲染出的规则行，按字段固定顺序排列，输出稳定可测；
//   - warnings：解析过程中跳过的未支持字段提示，调用方统一写日志，不中断输出；
//   - err：Content 整体非法时返回，调用方据此返回 500 或回退。
func RenderLines(content string) (lines []string, warnings []string, err error) {
	rules, err := parseRawRules(content)
	if err != nil {
		return nil, nil, err
	}

	lines = make([]string, 0, len(rules))
	for _, raw := range rules {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			// 单条规则不是对象（例如数组里混入了字符串），记录 warning 并跳过，不影响其它规则。
			warnings = append(warnings, "ruleset: skip non-object rule item")
			continue
		}

		lines = appendDomainLines(lines, fields, fieldDomain, LineRuleDomain)
		lines = appendDomainLines(lines, fields, fieldDomainSuffix, LineRuleDomainSuffix)
		lines = appendDomainLines(lines, fields, fieldDomainKeyword, LineRuleDomainKeyword)
		lines = appendDomainLines(lines, fields, fieldDomainRegex, LineRuleDomainRegex)
		lines = appendIPCIDRLines(lines, fields)
		lines = appendGEOIPLines(lines, fields)
		lines = appendProcessLines(lines, fields, fieldProcessName)
		lines = appendProcessLines(lines, fields, fieldProcessPath)

		// 记录未支持字段，便于后续按需扩展映射规则。
		for key := range fields {
			if _, ok := supportedFields[key]; !ok {
				warnings = append(warnings, "ruleset: unsupported field skipped: "+key)
			}
		}
	}
	return lines, warnings, nil
}

// CountLines 返回规则集内容展开为逐行规则后的规则条数，供生成链路判定是否值得改用远程 URL 引用。
// 计数口径与 RenderLines 一致（每个域名 / CIDR / 进程项各计一条），忽略 warning；
// Content 整体非法时返回 error，调用方据此回退到内联或跳过。
func CountLines(content string) (int, error) {
	lines, _, err := RenderLines(content)
	if err != nil {
		return 0, err
	}
	return len(lines), nil
}

// appendDomainLines 把某个字符串列表字段渲染为对应类型的规则行。
func appendDomainLines(lines []string, fields map[string]json.RawMessage, field string, ruleType LineRuleType) []string {
	for _, value := range stringValues(fields, field) {
		lines = append(lines, string(ruleType)+","+value)
	}
	return lines
}

// appendIPCIDRLines 把 ip_cidr 字段拆分为 IPv4（IP-CIDR）与 IPv6（IP-CIDR6）两类规则行。
func appendIPCIDRLines(lines []string, fields map[string]json.RawMessage) []string {
	for _, cidr := range stringValues(fields, fieldIPCIDR) {
		ruleType := LineRuleIPCIDR
		if ip, _, err := net.ParseCIDR(cidr); err == nil && ip.To4() == nil {
			ruleType = LineRuleIPCIDR6
		}
		lines = append(lines, string(ruleType)+","+cidr)
	}
	return lines
}

// appendGEOIPLines 把 geoip 字段渲染为大写国家/地区代码的 GEOIP 规则行，并附带 no-resolve 选项。
func appendGEOIPLines(lines []string, fields map[string]json.RawMessage) []string {
	for _, value := range stringValues(fields, fieldGEOIP) {
		lines = append(lines, string(LineRuleGEOIP)+","+strings.ToUpper(value)+","+geoipNoResolve)
	}
	return lines
}

// appendProcessLines 把 process_name / process_path 字段渲染为 PROCESS-NAME 规则行。
// 含空格的可执行文件路径会用双引号包裹，满足 Surge / Shadowrocket 对带空格值的引用要求。
func appendProcessLines(lines []string, fields map[string]json.RawMessage, field string) []string {
	for _, value := range stringValues(fields, field) {
		lines = append(lines, string(LineRuleProcessName)+","+QuoteValue(value))
	}
	return lines
}

// QuoteValue 在值包含空白字符或路径分隔符（/、\）时用双引号包裹，否则原样返回。
// Surge / Shadowrocket 规则行以逗号分隔：带空格的进程路径（如 "Lark Helper"）必须加引号才能被正确解析；
// 而像 /Applications/WeChat.app/Contents/MacOS/WeChat 这类不含空格的可执行文件全路径也统一加引号，
// 与带空格路径保持一致的书写形态，避免客户端把路径误当作进程名匹配。
func QuoteValue(value string) string {
	if strings.ContainsAny(value, " \t/\\") {
		return "\"" + value + "\""
	}
	return value
}

// stringValues 从规则字段里读取字符串数组，并裁剪空白、跳过空值。
// 字段缺失或不是字符串数组时返回空切片，保证调用方无需额外判空。
func stringValues(fields map[string]json.RawMessage, field string) []string {
	raw, ok := fields[field]
	if !ok {
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}

// BuildRuleSetURL 拼接规则集 open 接口的绝对访问地址，供生成链路把 local / inline 规则集
// 输出为远程 URL 引用。路径上只保留规则集 tag（PathEscape 转义），software / device / token
// 一律走 query 参数（QueryEscape 转义）；systemHost 先去掉尾部斜杠，确保特殊字符（空格、斜杠、
// 问号等）都能正确还原。
func BuildRuleSetURL(systemHost string, tag string, software entity.Software, device string, token string) string {
	host := strings.TrimRight(strings.TrimSpace(systemHost), "/")
	query := url.Values{}
	query.Set("software", string(software))
	query.Set("device", device)
	query.Set("token", token)

	var builder strings.Builder
	builder.WriteString(host)
	builder.WriteString("/open/rules/")
	builder.WriteString(url.PathEscape(tag))
	builder.WriteString("?")
	builder.WriteString(query.Encode())
	return builder.String()
}
