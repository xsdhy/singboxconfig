package surge

import (
	"encoding/json"
	"fmt"
	"net"
	"singboxconfig/convert/common"
	"singboxconfig/entity"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	// defaultPolicyName 是现有 sing-box 路由链路使用的兜底出站名称。
	defaultPolicyName = "general"
	// defaultURLTestURL 是节点分组未配置探测地址时的默认值。
	defaultURLTestURL = "https://www.gstatic.com/generate_204"
	// directPolicyName 是 Surge 内置的直连策略名称。
	directPolicyName = "DIRECT"
	// rejectPolicyName 是 Surge 内置的拒绝策略名称。
	rejectPolicyName = "REJECT"
	// systemDNSServer 是 Surge 使用系统 DNS 的保底写法。
	systemDNSServer = "system"
)

type surgeProxyProtocol string

const (
	// surgeProxySS 表示 Surge Shadowsocks 代理协议关键字。
	surgeProxySS surgeProxyProtocol = "ss"
	// surgeProxyTrojan 表示 Surge Trojan 代理协议关键字。
	surgeProxyTrojan surgeProxyProtocol = "trojan"
	// surgeProxyVMess 表示 Surge VMess 代理协议关键字。
	surgeProxyVMess surgeProxyProtocol = "vmess"
)

type surgeGroupType string

const (
	// surgeGroupSelect 表示 Surge 手动选择策略组。
	surgeGroupSelect surgeGroupType = "select"
	// surgeGroupURLTest 表示 Surge 延迟自动测试策略组。
	surgeGroupURLTest surgeGroupType = "url-test"
)

type surgeRuleType string

const (
	// surgeRuleDomain 表示域名精确匹配规则。
	surgeRuleDomain surgeRuleType = "DOMAIN"
	// surgeRuleDomainSuffix 表示域名后缀匹配规则。
	surgeRuleDomainSuffix surgeRuleType = "DOMAIN-SUFFIX"
	// surgeRuleDomainKeyword 表示域名关键字匹配规则。
	surgeRuleDomainKeyword surgeRuleType = "DOMAIN-KEYWORD"
	// surgeRuleDomainRegex 表示域名正则匹配规则。
	surgeRuleDomainRegex surgeRuleType = "DOMAIN-REGEX"
	// surgeRuleIPCIDR 表示 IPv4 CIDR 匹配规则。
	surgeRuleIPCIDR surgeRuleType = "IP-CIDR"
	// surgeRuleIPCIDR6 表示 IPv6 CIDR 匹配规则。
	surgeRuleIPCIDR6 surgeRuleType = "IP-CIDR6"
	// surgeRuleGEOIP 表示国家或地区 IP 库匹配规则。
	surgeRuleGEOIP surgeRuleType = "GEOIP"
	// surgeRuleRuleSet 表示远程或内联规则集引用。
	surgeRuleRuleSet surgeRuleType = "RULE-SET"
	// surgeRuleFinal 表示兜底规则。
	surgeRuleFinal surgeRuleType = "FINAL"
)

type renderContext struct {
	// proxyNames 记录 sing-box outbound tag 到 Surge 代理名称的映射。
	proxyNames map[string]string
	// proxyTags 按成功导出顺序记录 sing-box outbound tag，保证分组成员顺序稳定。
	proxyTags []string
	// groupNames 记录节点分组 tag 到 Surge 策略组名称的映射。
	groupNames map[string]string
	// warnings 收集可降级问题，统一写入日志，避免中断整体配置生成。
	warnings []string
}

// Render 将统一中间数据渲染为 Surge INI 风格配置文本。
// 入参全部来自现有生成链路：DNS、统一 Outbound、NodeGroup 与 RuleSet，
// 函数内部只做纯转换，便于对协议映射、分组一致性和规则展开做单元测试。
func Render(deviceCode string, dns entity.SingDNS, outbounds []entity.SingBoxOut, groups []*entity.NodeGroup, ruleSets []*entity.RuleSet) string {
	ctx := &renderContext{
		proxyNames: make(map[string]string),
		groupNames: make(map[string]string),
	}

	proxyLines := renderProxySection(ctx, outbounds)
	groupLines := renderProxyGroupSection(ctx, groups)
	ruleLines := renderRuleSection(ctx, deviceCode, ruleSets)
	ctx.flushWarnings()

	sections := [][]string{
		renderGeneralSection(dns),
		withSectionHeader("[Proxy]", proxyLines),
		withSectionHeader("[Proxy Group]", groupLines),
		withSectionHeader("[Rule]", ruleLines),
	}

	var builder strings.Builder
	for index, section := range sections {
		if index > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(strings.Join(section, "\n"))
		builder.WriteString("\n")
	}
	return builder.String()
}

func renderGeneralSection(dns entity.SingDNS) []string {
	return []string{
		"[General]",
		"loglevel = notify",
		"ipv6 = true",
		"dns-server = " + strings.Join(resolveDNSServers(dns), ", "),
	}
}

func resolveDNSServers(dns entity.SingDNS) []string {
	servers := make([]string, 0, len(dns.Servers))
	seen := make(map[string]struct{}, len(dns.Servers))
	for _, server := range dns.Servers {
		address := strings.TrimSpace(server.Address)
		if address == "" || strings.HasPrefix(address, "rcode://") {
			continue
		}
		if _, ok := seen[address]; ok {
			continue
		}
		seen[address] = struct{}{}
		servers = append(servers, address)
	}
	if len(servers) == 0 {
		return []string{systemDNSServer}
	}
	return servers
}

func renderProxySection(ctx *renderContext, outbounds []entity.SingBoxOut) []string {
	lines := make([]string, 0, len(outbounds))
	usedNames := make(map[string]struct{}, len(outbounds))
	for _, outbound := range outbounds {
		line, name, ok := renderProxyLine(ctx, outbound)
		if !ok {
			continue
		}
		if _, exists := usedNames[name]; exists {
			ctx.warnf("Surge proxy name duplicated, skip outbound: tag=%s name=%s", outbound.Tag, name)
			continue
		}
		usedNames[name] = struct{}{}
		ctx.proxyNames[outbound.Tag] = name
		ctx.proxyTags = append(ctx.proxyTags, outbound.Tag)
		lines = append(lines, line)
	}
	return lines
}

func renderProxyLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	protocolType := entity.OutboundProtocol(outbound.Type)
	switch protocolType {
	case entity.OutboundProtocolShadowsocks:
		return renderShadowsocksLine(ctx, outbound)
	case entity.OutboundProtocolTrojan:
		return renderTrojanLine(ctx, outbound)
	case entity.OutboundProtocolVMess:
		return renderVMessLine(ctx, outbound)
	case entity.OutboundProtocolVLESS, entity.OutboundProtocolHysteria, entity.OutboundProtocolHysteria2, entity.OutboundProtocolTUIC:
		ctx.warnf("Surge unsupported outbound protocol, skip: tag=%s type=%s", outbound.Tag, outbound.Type)
		return "", "", false
	default:
		if outbound.Type != "" && outbound.Type != string(entity.NodeGroupTypeSelector) && outbound.Type != string(entity.NodeGroupTypeURLTest) && outbound.Type != string(entity.OutboundProtocolDirect) {
			ctx.warnf("Surge unrecognized outbound protocol, skip: tag=%s type=%s", outbound.Tag, outbound.Type)
		}
		return "", "", false
	}
}

func renderShadowsocksLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	if outbound.Method == "" || outbound.Password == "" {
		ctx.warnf("Surge shadowsocks outbound missing method or password, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	parts := []string{
		name + " = " + string(surgeProxySS),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue("encrypt-method", outbound.Method),
		keyValue("password", outbound.Password),
		keyValue("udp-relay", "true"),
	}
	appendShadowsocksPluginParameters(&parts, outbound)
	return strings.Join(parts, ", "), name, true
}

func renderTrojanLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	if outbound.Password == "" {
		ctx.warnf("Surge trojan outbound missing password, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	parts := []string{
		name + " = " + string(surgeProxyTrojan),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue("password", outbound.Password),
		keyValue("udp-relay", "true"),
	}
	appendTLSParameters(&parts, outbound.TLS)
	return strings.Join(parts, ", "), name, true
}

func renderVMessLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	if outbound.UUID == "" {
		ctx.warnf("Surge vmess outbound missing uuid, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	ctx.warnf("Surge vmess outbound is best-effort and only suitable for Surge 4+: tag=%s", outbound.Tag)
	parts := []string{
		name + " = " + string(surgeProxyVMess),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue("username", outbound.UUID),
	}
	if outbound.Security != "" && outbound.Security != "auto" {
		parts = append(parts, keyValue("encrypt-method", outbound.Security))
	}
	appendTLSParameters(&parts, outbound.TLS)
	appendVMessTransportParameters(&parts, outbound.Transport)
	return strings.Join(parts, ", "), name, true
}

func validateProxyBasics(ctx *renderContext, outbound entity.SingBoxOut) (string, bool) {
	name := policyName(outbound.Tag)
	if name == "" {
		ctx.warnf("Surge outbound missing tag, skip: type=%s server=%s", outbound.Type, outbound.Server)
		return "", false
	}
	if strings.TrimSpace(outbound.Server) == "" || outbound.ServerPort <= 0 {
		ctx.warnf("Surge outbound missing server or server_port, skip: tag=%s", outbound.Tag)
		return "", false
	}
	return name, true
}

func appendTLSParameters(parts *[]string, tls *entity.SingTLS) {
	if tls == nil || !tls.Enabled {
		return
	}
	*parts = append(*parts, keyValue("tls", "true"))
	if tls.ServerName != "" {
		*parts = append(*parts, keyValue("sni", tls.ServerName))
	}
	if tls.Insecure {
		*parts = append(*parts, keyValue("skip-cert-verify", "true"))
	}
	if len(tls.Alpn) > 0 {
		*parts = append(*parts, keyValue("alpn", strings.Join(tls.Alpn, "|")))
	}
}

func appendShadowsocksPluginParameters(parts *[]string, outbound entity.SingBoxOut) {
	pluginOptions := parseSemicolonOptions(outbound.PluginOpts)
	if obfs := firstNonEmpty(pluginOptions["obfs"], outbound.Plugin); obfs == "http" || obfs == "tls" {
		*parts = append(*parts, keyValue("obfs", obfs))
	}
	if obfsHost := firstNonEmpty(pluginOptions["obfs-host"], outbound.ObfsParam); obfsHost != "" {
		*parts = append(*parts, keyValue("obfs-host", obfsHost))
	}
	if obfsURI := pluginOptions["obfs-uri"]; obfsURI != "" {
		*parts = append(*parts, keyValue("obfs-uri", obfsURI))
	}
}

func appendVMessTransportParameters(parts *[]string, transport *entity.SingTransport) {
	if transport == nil {
		return
	}
	if transport.Type == "ws" {
		*parts = append(*parts, keyValue("ws", "true"))
		if transport.Path != "" {
			*parts = append(*parts, keyValue("ws-path", transport.Path))
		}
		if host := firstHeaderValue(transport.Headers, "Host"); host != "" {
			*parts = append(*parts, keyValue("ws-headers", "Host:"+host))
		}
	}
}

func renderProxyGroupSection(ctx *renderContext, groups []*entity.NodeGroup) []string {
	if len(groups) == 0 {
		return nil
	}
	tags := exportedProxyTags(ctx)
	lines := make([]string, 0, len(groups))
	for _, group := range groups {
		if group == nil || group.Tag == "" {
			continue
		}
		members := common.FilterOutboundGroupTags(group, tags)
		if len(members) == 0 {
			continue
		}
		groupName := policyName(group.Tag)
		ctx.groupNames[group.Tag] = groupName

		lineMembers := make([]string, 0, len(members)+4)
		lineMembers = append(lineMembers, groupName+" = "+string(resolveSurgeGroupType(group.GroupType)))
		for _, tag := range members {
			lineMembers = append(lineMembers, ctx.proxyNames[tag])
		}
		if entity.NodeGroupType(group.GroupType) == entity.NodeGroupTypeURLTest {
			testURL := strings.TrimSpace(group.TestURL)
			if testURL == "" {
				testURL = defaultURLTestURL
			}
			lineMembers = append(lineMembers, keyValue("url", testURL), keyValue("interval", "600"), keyValue("tolerance", "50"))
		}
		lines = append(lines, strings.Join(lineMembers, ", "))
	}
	return lines
}

func exportedProxyTags(ctx *renderContext) []string {
	return append([]string(nil), ctx.proxyTags...)
}

func resolveSurgeGroupType(groupType string) surgeGroupType {
	if entity.NodeGroupType(groupType) == entity.NodeGroupTypeURLTest {
		return surgeGroupURLTest
	}
	return surgeGroupSelect
}

func renderRuleSection(ctx *renderContext, deviceCode string, ruleSets []*entity.RuleSet) []string {
	sorted := append([]*entity.RuleSet(nil), ruleSets...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i] == nil || sorted[j] == nil {
			return sorted[j] != nil
		}
		return sorted[i].Sort < sorted[j].Sort
	})

	lines := make([]string, 0, len(sorted)+1)
	for _, ruleSet := range sorted {
		if ruleSet == nil || !isRuleSetVisibleForDevice(ruleSet, deviceCode) {
			continue
		}
		policy := ctx.policyReference(ruleSet.Outbound)
		if entity.RuleSetType(ruleSet.RuleSetType) == entity.RuleSetTypeRemote {
			if strings.TrimSpace(ruleSet.URL) == "" {
				ctx.warnf("Surge remote ruleset missing url, skip: tag=%s", ruleSet.Tag)
				continue
			}
			lines = append(lines, strings.Join([]string{string(surgeRuleRuleSet), ruleSet.URL, policy}, ","))
			continue
		}
		lines = append(lines, expandLocalRuleSet(ctx, ruleSet, policy)...)
	}
	lines = append(lines, strings.Join([]string{string(surgeRuleFinal), ctx.policyReference(defaultPolicyName)}, ","))
	return lines
}

func isRuleSetVisibleForDevice(ruleSet *entity.RuleSet, deviceCode string) bool {
	return strings.TrimSpace(ruleSet.AbleDevices) == "" || strings.Contains(ruleSet.AbleDevices, deviceCode)
}

type singBoxRuleSetContent struct {
	// Rules 是 sing-box source 规则集中的规则数组。
	Rules []singBoxRule `json:"rules"`
}

type singBoxRule struct {
	// Domain 表示精确域名匹配列表。
	Domain []string `json:"domain"`
	// DomainSuffix 表示域名后缀匹配列表。
	DomainSuffix []string `json:"domain_suffix"`
	// DomainKeyword 表示域名关键字匹配列表。
	DomainKeyword []string `json:"domain_keyword"`
	// DomainRegex 表示域名正则匹配列表。
	DomainRegex []string `json:"domain_regex"`
	// IPCIDR 表示 IPv4 或 IPv6 CIDR 匹配列表。
	IPCIDR []string `json:"ip_cidr"`
	// GEOIP 表示国家或地区 IP 库匹配列表。
	GEOIP []string `json:"geoip"`
}

func expandLocalRuleSet(ctx *renderContext, ruleSet *entity.RuleSet, policy string) []string {
	rules, err := parseLocalRules(ruleSet.Content)
	if err != nil {
		ctx.warnf("Surge local ruleset content invalid, skip: tag=%s err=%v", ruleSet.Tag, err)
		return nil
	}
	lines := make([]string, 0, len(rules))
	for _, rule := range rules {
		lines = appendRules(lines, surgeRuleDomain, rule.Domain, policy)
		lines = appendRules(lines, surgeRuleDomainSuffix, rule.DomainSuffix, policy)
		lines = appendRules(lines, surgeRuleDomainKeyword, rule.DomainKeyword, policy)
		lines = appendRules(lines, surgeRuleDomainRegex, rule.DomainRegex, policy)
		for _, cidr := range rule.IPCIDR {
			ruleType := surgeRuleIPCIDR
			if ip, _, err := net.ParseCIDR(cidr); err == nil && ip.To4() == nil {
				ruleType = surgeRuleIPCIDR6
			}
			lines = append(lines, strings.Join([]string{string(ruleType), cidr, policy}, ","))
		}
		for _, geoip := range rule.GEOIP {
			lines = append(lines, strings.Join([]string{string(surgeRuleGEOIP), strings.ToUpper(geoip), policy}, ","))
		}
	}
	return lines
}

func parseLocalRules(content string) ([]singBoxRule, error) {
	var ruleSetContent singBoxRuleSetContent
	if err := json.Unmarshal([]byte(content), &ruleSetContent); err == nil && len(ruleSetContent.Rules) > 0 {
		return ruleSetContent.Rules, nil
	}

	var rules []singBoxRule
	if err := json.Unmarshal([]byte(content), &rules); err == nil {
		return rules, nil
	}

	var singleRule singBoxRule
	if err := json.Unmarshal([]byte(content), &singleRule); err != nil {
		return nil, err
	}
	return []singBoxRule{singleRule}, nil
}

func appendRules(lines []string, ruleType surgeRuleType, values []string, policy string) []string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		lines = append(lines, strings.Join([]string{string(ruleType), value, policy}, ","))
	}
	return lines
}

func (ctx *renderContext) policyReference(tag string) string {
	trimmed := strings.TrimSpace(tag)
	if trimmed == "" {
		return defaultPolicyName
	}
	switch strings.ToLower(trimmed) {
	case "direct":
		return directPolicyName
	case "reject":
		return rejectPolicyName
	default:
		if name, ok := ctx.proxyNames[trimmed]; ok {
			return name
		}
		if name, ok := ctx.groupNames[trimmed]; ok {
			return name
		}
		return policyName(trimmed)
	}
}

func policyName(raw string) string {
	name := strings.TrimSpace(raw)
	name = strings.ReplaceAll(name, ",", " ")
	name = strings.ReplaceAll(name, "=", " ")
	return strings.Join(strings.Fields(name), " ")
}

func keyValue(key string, value string) string {
	return key + "=" + encodeValue(value)
}

func encodeValue(value string) string {
	if strings.ContainsAny(value, ",#\"") {
		escaped := strings.ReplaceAll(value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return "\"" + escaped + "\""
	}
	return value
}

func firstHeaderValue(headers map[string][]string, key string) string {
	for headerKey, values := range headers {
		if strings.EqualFold(headerKey, key) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func parseSemicolonOptions(raw string) map[string]string {
	result := make(map[string]string)
	for _, part := range strings.Split(raw, ";") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" {
			result[key] = value
		}
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func withSectionHeader(header string, lines []string) []string {
	if len(lines) == 0 {
		return []string{header}
	}
	result := make([]string, 0, len(lines)+1)
	result = append(result, header)
	result = append(result, lines...)
	return result
}

func (ctx *renderContext) warnf(format string, args ...any) {
	ctx.warnings = append(ctx.warnings, fmt.Sprintf(format, args...))
}

func (ctx *renderContext) flushWarnings() {
	for _, warning := range ctx.warnings {
		logrus.Warn(warning)
	}
}
