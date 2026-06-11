package surge

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"singboxconfig/convert/common"
	"singboxconfig/convert/ruleset"
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
)

type surgeProxyProtocol string

const (
	// surgeProxySS 表示 Surge Shadowsocks 代理协议关键字。
	surgeProxySS surgeProxyProtocol = "ss"
	// surgeProxyTrojan 表示 Surge Trojan 代理协议关键字。
	surgeProxyTrojan surgeProxyProtocol = "trojan"
	// surgeProxyVMess 表示 Surge VMess 代理协议关键字。
	surgeProxyVMess surgeProxyProtocol = "vmess"
	// surgeProxyHTTP 表示 Surge 明文 HTTP 代理协议关键字。
	surgeProxyHTTP surgeProxyProtocol = "http"
	// surgeProxyHTTPS 表示 Surge TLS HTTP 代理协议关键字。
	surgeProxyHTTPS surgeProxyProtocol = "https"
	// surgeProxySocks5 表示 Surge 明文 SOCKS5 代理协议关键字。
	surgeProxySocks5 surgeProxyProtocol = "socks5"
	// surgeProxySocks5TLS 表示 Surge TLS SOCKS5 代理协议关键字。
	surgeProxySocks5TLS surgeProxyProtocol = "socks5-tls"
	// surgeProxyWireGuard 表示 Surge WireGuard 代理协议关键字。
	surgeProxyWireGuard surgeProxyProtocol = "wireguard"
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
	// surgeRuleProcessName 表示进程匹配规则，值可为进程名或可执行文件全路径。
	surgeRuleProcessName surgeRuleType = "PROCESS-NAME"
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
	// subscriptionGroupNames 按导出顺序记录订阅策略组名称，供节点分组通过 include-other-group 引用。
	subscriptionGroupNames []string
	// wireGuardSections 按导出顺序记录 WireGuard 独立配置段文本，追加到配置末尾。
	wireGuardSections []string
	// warnings 收集可降级问题，统一写入日志，避免中断整体配置生成。
	warnings []string
}

// Render 将统一中间数据渲染为 Surge INI 风格配置文本。
// 入参全部来自现有生成链路：统一 Outbound、WireGuard endpoint、订阅源、NodeGroup 与 RuleSet，
// 函数内部只做纯转换，便于对协议映射、分组一致性和规则展开做单元测试。
//
// outbounds 只应包含手工维护节点：订阅节点不再逐条展开为 [Proxy] 行，
// 而是为每个订阅生成一个携带 policy-path=<订阅地址> 的策略组，由 Surge 自行拉取订阅，
// 节点分组再通过 include-other-group 引用订阅策略组，并用 policy-regex-filter
// 还原 Include/Exclude 关键字过滤语义。
//
// deviceToken 与 systemHost 用于把 local / inline 规则集改为单行 RULE-SET,<url>,<policy> 引用：
// systemHost 非空且内容可解析时输出 URL 引用，否则回退到逐条展开，保证旧部署零配置可用。
func Render(deviceCode string, deviceToken string, systemHost string, outbounds []entity.SingBoxOut, endpoints []entity.SingEndpointWireguard, subscribes []*entity.Subscribe, groups []*entity.NodeGroup, ruleSets []*entity.RuleSet) string {
	ctx := &renderContext{
		proxyNames: make(map[string]string),
		groupNames: make(map[string]string),
	}

	proxyLines := renderProxySection(ctx, outbounds)
	proxyLines = append(proxyLines, renderWireGuardProxyLines(ctx, endpoints)...)
	subscriptionGroupLines := renderSubscriptionGroupSection(ctx, deviceCode, subscribes)
	groupLines := append(subscriptionGroupLines, renderProxyGroupSection(ctx, deviceCode, groups)...)
	ruleLines := renderRuleSection(ctx, deviceCode, deviceToken, systemHost, ruleSets)
	ctx.flushWarnings()

	sections := [][]string{
		renderGeneralSection(),
		withSectionHeader("[Proxy]", proxyLines),
		withSectionHeader("[Proxy Group]", groupLines),
		withSectionHeader("[Rule]", ruleLines),
	}
	for _, wireGuardSection := range ctx.wireGuardSections {
		sections = append(sections, []string{wireGuardSection})
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

func renderGeneralSection() []string {
	return []string{
		"[General]",
		"loglevel = notify",
		"ipv6 = false",
	}
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

// renderWireGuardProxyLines 把 sing-box 的 WireGuard endpoint 转换为 Surge 代理。
// 每个 endpoint 产出一条 [Proxy] 引用行与一段独立的 [WireGuard <Section>] 配置，
// 同时把 endpoint tag 注册到 proxyNames/proxyTags，使其能被策略组和规则正常引用。
func renderWireGuardProxyLines(ctx *renderContext, endpoints []entity.SingEndpointWireguard) []string {
	lines := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		name := policyName(endpoint.Tag)
		if name == "" {
			ctx.warnf("Surge wireguard endpoint missing tag, skip")
			continue
		}
		if _, exists := ctx.proxyNames[endpoint.Tag]; exists {
			ctx.warnf("Surge proxy name duplicated, skip wireguard endpoint: tag=%s name=%s", endpoint.Tag, name)
			continue
		}
		if strings.TrimSpace(endpoint.PrivateKey) == "" {
			ctx.warnf("Surge wireguard endpoint missing private_key, skip: tag=%s", endpoint.Tag)
			continue
		}
		peerLines := wireGuardPeerLines(ctx, endpoint)
		if len(peerLines) == 0 {
			ctx.warnf("Surge wireguard endpoint has no usable peer, skip: tag=%s", endpoint.Tag)
			continue
		}

		sectionName := name
		ctx.proxyNames[endpoint.Tag] = name
		ctx.proxyTags = append(ctx.proxyTags, endpoint.Tag)
		lines = append(lines, name+" = "+string(surgeProxyWireGuard)+", "+keyValue("section-name", sectionName))
		ctx.wireGuardSections = append(ctx.wireGuardSections, renderWireGuardSection(sectionName, endpoint, peerLines))
	}
	return lines
}

func renderWireGuardSection(sectionName string, endpoint entity.SingEndpointWireguard, peerLines []string) string {
	sectionLines := []string{"[WireGuard " + sectionName + "]"}
	sectionLines = append(sectionLines, "private-key = "+endpoint.PrivateKey)

	ipv4, ipv6 := splitWireGuardAddresses(endpoint.Address)
	if ipv4 != "" {
		sectionLines = append(sectionLines, "self-ip = "+ipv4)
	}
	if ipv6 != "" {
		sectionLines = append(sectionLines, "self-ip-v6 = "+ipv6)
	}
	if endpoint.MTU > 0 {
		sectionLines = append(sectionLines, "mtu = "+strconv.Itoa(endpoint.MTU))
	}
	sectionLines = append(sectionLines, peerLines...)
	return strings.Join(sectionLines, "\n")
}

func wireGuardPeerLines(ctx *renderContext, endpoint entity.SingEndpointWireguard) []string {
	lines := make([]string, 0, len(endpoint.Peers))
	for _, peer := range endpoint.Peers {
		if strings.TrimSpace(peer.PublicKey) == "" {
			ctx.warnf("Surge wireguard peer missing public_key, skip: tag=%s", endpoint.Tag)
			continue
		}
		if strings.TrimSpace(peer.Address) == "" || peer.Port <= 0 {
			ctx.warnf("Surge wireguard peer missing endpoint address or port, skip: tag=%s", endpoint.Tag)
			continue
		}
		allowedIPs := peer.AllowedIps
		if len(allowedIPs) == 0 {
			allowedIPs = []string{"0.0.0.0/0", "::/0"}
		}
		// 多个 allowed-ips 用逗号拼接后，必须整体加引号，
		// 否则内部逗号会被 Surge 当成 peer 字段分隔符而解析错误。
		allowedValue := strings.Join(allowedIPs, ", ")
		if len(allowedIPs) > 1 {
			allowedValue = "\"" + allowedValue + "\""
		}
		fields := []string{
			"public-key = " + peer.PublicKey,
			"allowed-ips = " + allowedValue,
			"endpoint = " + net.JoinHostPort(peer.Address, strconv.Itoa(peer.Port)),
		}
		if strings.TrimSpace(peer.PreSharedKey) != "" {
			fields = append(fields, "preshared-key = "+peer.PreSharedKey)
		}
		if peer.PersistentKeepaliveInterval > 0 {
			fields = append(fields, "keepalive = "+strconv.Itoa(peer.PersistentKeepaliveInterval))
		}
		lines = append(lines, "peer = ("+strings.Join(fields, ", ")+")")
	}
	return lines
}

// splitWireGuardAddresses 把 sing-box 的客户端地址列表拆成 IPv4 与 IPv6，
// 取首个匹配项；Surge 的 self-ip 字段只接受纯 IP，需去掉 CIDR 前缀。
func splitWireGuardAddresses(addresses []string) (string, string) {
	var ipv4, ipv6 string
	for _, addr := range addresses {
		host := strings.TrimSpace(addr)
		if host == "" {
			continue
		}
		if idx := strings.Index(host, "/"); idx >= 0 {
			host = host[:idx]
		}
		parsed := net.ParseIP(host)
		if parsed != nil && parsed.To4() == nil {
			if ipv6 == "" {
				ipv6 = host
			}
			continue
		}
		if ipv4 == "" {
			ipv4 = host
		}
	}
	return ipv4, ipv6
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
	case entity.OutboundProtocolHTTP:
		return renderHTTPLine(ctx, outbound)
	case entity.OutboundProtocolSocks:
		return renderSocksLine(ctx, outbound)
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

func renderHTTPLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	// sing-box 通过 tls.enabled 区分明文 HTTP 与 HTTPS，Surge 中对应 http / https 两种关键字。
	protocol := surgeProxyHTTP
	if outbound.TLS != nil && outbound.TLS.Enabled {
		protocol = surgeProxyHTTPS
	}
	parts := []string{
		name + " = " + string(protocol),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
	}
	if outbound.Username != "" {
		parts = append(parts, keyValue("username", outbound.Username))
	}
	if outbound.Password != "" {
		parts = append(parts, keyValue("password", outbound.Password))
	}
	appendTLSParameters(&parts, outbound.TLS)
	return strings.Join(parts, ", "), name, true
}

func renderSocksLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	// sing-box 的 socks 出站不带 version 时默认 SOCKS5；Surge 不支持 SOCKS4，仅在显式声明旧版本时跳过。
	if outbound.Version != 0 && outbound.Version != 5 {
		ctx.warnf("Surge only supports socks5, skip: tag=%s version=%d", outbound.Tag, outbound.Version)
		return "", "", false
	}
	protocol := surgeProxySocks5
	if outbound.TLS != nil && outbound.TLS.Enabled {
		protocol = surgeProxySocks5TLS
	}
	parts := []string{
		name + " = " + string(protocol),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
	}
	// Surge 的 socks5 用户名密码为位置参数，必须成对出现。
	if outbound.Username != "" && outbound.Password != "" {
		parts = append(parts, outbound.Username, outbound.Password)
	} else if outbound.Username != "" || outbound.Password != "" {
		ctx.warnf("Surge socks5 outbound requires both username and password, auth dropped: tag=%s", outbound.Tag)
	}
	parts = append(parts, keyValue("udp-relay", "true"))
	if outbound.TLS != nil && outbound.TLS.Enabled {
		if outbound.TLS.ServerName != "" {
			parts = append(parts, keyValue("sni", outbound.TLS.ServerName))
		}
		if outbound.TLS.Insecure {
			parts = append(parts, keyValue("skip-cert-verify", "true"))
		}
	}
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

func renderProxyGroupSection(ctx *renderContext, deviceCode string, groups []*entity.NodeGroup) []string {
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
		// 没有任何手工节点命中且没有订阅策略组可引用时，跳过空分组。
		if len(members) == 0 && len(ctx.subscriptionGroupNames) == 0 {
			continue
		}
		groupName := policyName(group.Tag)
		ctx.groupNames[group.Tag] = groupName

		// 按设备解析最终分组类型，确保 url-test 关键字与 url/interval/tolerance 参数基于同一次判定，
		// 避免设备覆盖为 urltest 时类型变了但探测参数缺失。
		groupType := common.ResolveGroupType(group, deviceCode)

		lineMembers := make([]string, 0, len(members)+6)
		lineMembers = append(lineMembers, groupName+" = "+string(resolveSurgeGroupType(groupType)))
		for _, tag := range members {
			lineMembers = append(lineMembers, ctx.proxyNames[tag])
		}
		// 订阅节点不再展开成 [Proxy] 行，这里通过 include-other-group 引用订阅策略组，
		// 并用 policy-regex-filter 还原 Include/Exclude 关键字过滤语义。
		if len(ctx.subscriptionGroupNames) > 0 {
			lineMembers = append(lineMembers, keyValue("include-other-group", strings.Join(ctx.subscriptionGroupNames, ",")))
			if filter := buildPolicyRegexFilter(group.Include, group.Exclude); filter != "" {
				lineMembers = append(lineMembers, keyValue("policy-regex-filter", filter))
			}
		}
		if groupType == entity.NodeGroupTypeURLTest {
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

// renderSubscriptionGroupSection 把每个订阅源渲染为一个携带 policy-path 的 select 策略组，
// 由 Surge 客户端自行拉取订阅地址，服务端不再展开订阅节点。
// 订阅的缓存时长（分钟）映射为 Surge 的 update-interval（秒）。
func renderSubscriptionGroupSection(ctx *renderContext, deviceCode string, subscribes []*entity.Subscribe) []string {
	lines := make([]string, 0, len(subscribes))
	for _, subscribe := range subscribes {
		if subscribe == nil || !subscribe.Status || !isSubscribeVisibleForDevice(subscribe, deviceCode) {
			continue
		}
		if strings.TrimSpace(subscribe.URL) == "" {
			ctx.warnf("Surge subscription missing url, skip: name=%s", subscribe.Name)
			continue
		}
		name := policyName(subscribe.Name)
		if name == "" {
			ctx.warnf("Surge subscription missing name, skip: url=%s", subscribe.URL)
			continue
		}
		if _, exists := ctx.proxyNames[subscribe.Name]; exists {
			ctx.warnf("Surge subscription name conflicts with proxy, skip: name=%s", name)
			continue
		}
		parts := []string{
			name + " = " + string(surgeGroupSelect),
			keyValue("policy-path", subscribe.URL),
		}
		if subscribe.OutboundCacheDuration > 0 {
			parts = append(parts, keyValue("update-interval", strconv.Itoa(subscribe.OutboundCacheDuration*60)))
		}
		ctx.subscriptionGroupNames = append(ctx.subscriptionGroupNames, name)
		lines = append(lines, strings.Join(parts, ", "))
	}
	return lines
}

// buildPolicyRegexFilter 把节点分组的 Include/Exclude 关键字翻译成 Surge 的 policy-regex-filter 正则。
// Include 关键字之间为“或”关系，Exclude 通过负向先行断言剔除；关键字本身做正则转义，
// 与 FilterOutboundGroupTags 的子串匹配语义保持一致。两者皆空时返回空串表示不过滤。
func buildPolicyRegexFilter(include string, exclude string) string {
	includes := quoteKeywords(include)
	excludes := quoteKeywords(exclude)
	switch {
	case len(includes) == 0 && len(excludes) == 0:
		return ""
	case len(excludes) == 0:
		return "(" + strings.Join(includes, "|") + ")"
	case len(includes) == 0:
		return "^(?!.*(" + strings.Join(excludes, "|") + "))"
	default:
		return "^(?!.*(" + strings.Join(excludes, "|") + ")).*(" + strings.Join(includes, "|") + ")"
	}
}

// quoteKeywords 拆分逗号分隔关键字并做正则转义，忽略空白项。
func quoteKeywords(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		result = append(result, regexp.QuoteMeta(item))
	}
	return result
}

func exportedProxyTags(ctx *renderContext) []string {
	return append([]string(nil), ctx.proxyTags...)
}

func resolveSurgeGroupType(groupType entity.NodeGroupType) surgeGroupType {
	if groupType == entity.NodeGroupTypeURLTest {
		return surgeGroupURLTest
	}
	return surgeGroupSelect
}

func renderRuleSection(ctx *renderContext, deviceCode string, deviceToken string, systemHost string, ruleSets []*entity.RuleSet) []string {
	sorted := append([]*entity.RuleSet(nil), ruleSets...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i] == nil || sorted[j] == nil {
			return sorted[j] != nil
		}
		return sorted[i].Sort < sorted[j].Sort
	})

	host := strings.TrimSpace(systemHost)
	lines := make([]string, 0, len(sorted)+1)
	for _, ruleSet := range sorted {
		if ruleSet == nil || !isRuleSetVisibleForDevice(ruleSet, deviceCode) {
			continue
		}
		// 规则引用的出站/策略组在当前设备配置中不存在时跳过该条规则，避免生成悬空策略引用。
		policy, ok := ctx.resolvePolicy(ruleSet.Outbound)
		if !ok {
			ctx.warnf("Surge ruleset outbound not found in proxies or groups, skip rule: tag=%s outbound=%s", ruleSet.Tag, ruleSet.Outbound)
			continue
		}
		if entity.RuleSetType(ruleSet.RuleSetType) == entity.RuleSetTypeRemote {
			if strings.TrimSpace(ruleSet.URL) == "" {
				ctx.warnf("Surge remote ruleset missing url, skip: tag=%s", ruleSet.Tag)
				continue
			}
			lines = append(lines, strings.Join([]string{string(surgeRuleRuleSet), ruleSet.URL, policy}, ","))
			continue
		}

		// local / inline：系统 Host 可用且内容可解析时，输出单行 RULE-SET,<url>,<policy> 引用本服务 open 接口。
		// 但规则条数少于 ruleset.InlineThreshold 时直接内联展开，避免为极少量规则额外引入一次远程请求。
		if host != "" {
			if count, err := ruleset.CountLines(ruleSet.Content); err == nil {
				if count >= ruleset.InlineThreshold {
					rulesetURL := ruleset.BuildRuleSetURL(host, ruleSet.Tag, entity.SoftwareSurge, deviceCode, deviceToken)
					lines = append(lines, strings.Join([]string{string(surgeRuleRuleSet), rulesetURL, policy}, ","))
					continue
				}
			} else {
				// 内容非法时不生成指向坏内容的 URL，回退到逐条展开（展开同样会跳过坏内容并记录 warning）。
				ctx.warnf("Surge local ruleset content invalid, fallback to expand: tag=%s", ruleSet.Tag)
			}
		}
		lines = append(lines, expandLocalRuleSet(ctx, ruleSet, policy)...)
	}
	lines = append(lines, strings.Join([]string{string(surgeRuleFinal), ctx.policyReference(defaultPolicyName)}, ","))
	return lines
}

func isRuleSetVisibleForDevice(ruleSet *entity.RuleSet, deviceCode string) bool {
	return strings.TrimSpace(ruleSet.AbleDevices) == "" || strings.Contains(ruleSet.AbleDevices, deviceCode)
}

// isSubscribeVisibleForDevice 复用订阅源的逗号分隔设备可见性规则。
func isSubscribeVisibleForDevice(subscribe *entity.Subscribe, deviceCode string) bool {
	if strings.TrimSpace(subscribe.VisibleDevices) == "" {
		return true
	}
	for _, item := range strings.Split(subscribe.VisibleDevices, ",") {
		if strings.TrimSpace(item) == deviceCode {
			return true
		}
	}
	return false
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
	// ProcessName 表示进程名匹配列表（如 WeChat）。
	ProcessName []string `json:"process_name"`
	// ProcessPath 表示进程可执行文件全路径匹配列表（如 /Applications/WeChat.app/Contents/MacOS/WeChat）。
	ProcessPath []string `json:"process_path"`
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
			lines = append(lines, strings.Join([]string{string(surgeRuleGEOIP), strings.ToUpper(geoip), policy, "no-resolve"}, ","))
		}
		lines = appendProcessRules(lines, rule.ProcessName, policy)
		lines = appendProcessRules(lines, rule.ProcessPath, policy)
	}
	return lines
}

// appendProcessRules 把进程名 / 进程路径渲染为 PROCESS-NAME 规则行，带空格的值用双引号包裹。
func appendProcessRules(lines []string, values []string, policy string) []string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		lines = append(lines, strings.Join([]string{string(surgeRuleProcessName), ruleset.QuoteValue(value), policy}, ","))
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
	name, _ := ctx.resolvePolicy(tag)
	return name
}

// resolvePolicy 把规则出站标签解析为 Surge 策略名称，并返回该策略是否真实存在。
// 空标签与内置 DIRECT / REJECT 视为始终存在；其余标签必须命中已导出的代理或策略组，
// 否则 ok=false，供普通规则据此跳过悬空引用。FINAL 兜底通过 policyReference 忽略 ok。
func (ctx *renderContext) resolvePolicy(tag string) (string, bool) {
	trimmed := strings.TrimSpace(tag)
	if trimmed == "" {
		return defaultPolicyName, true
	}
	switch strings.ToLower(trimmed) {
	case "direct":
		return directPolicyName, true
	case "reject":
		return rejectPolicyName, true
	default:
		if name, ok := ctx.proxyNames[trimmed]; ok {
			return name, true
		}
		if name, ok := ctx.groupNames[trimmed]; ok {
			return name, true
		}
		return policyName(trimmed), false
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
