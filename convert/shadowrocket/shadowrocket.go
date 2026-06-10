package shadowrocket

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
	// defaultPolicyName 是现有路由规则中常用的默认策略组名称。
	defaultPolicyName = "general"
	// defaultURLTestURL 是节点分组未配置探测地址时的默认测试地址。
	defaultURLTestURL = "https://www.gstatic.com/generate_204"
	// directPolicyName 是 Shadowrocket 内置直连策略名称。
	directPolicyName = "DIRECT"
	// rejectPolicyName 是 Shadowrocket 内置拒绝策略名称。
	rejectPolicyName = "REJECT"
	// systemDNSServer 是 Shadowrocket 使用系统 DNS 的兜底写法。
	systemDNSServer = "system"
)

type shadowrocketSection string

const (
	// sectionGeneral 表示基础配置段。
	sectionGeneral shadowrocketSection = "[General]"
	// sectionProxy 表示代理节点段。
	sectionProxy shadowrocketSection = "[Proxy]"
	// sectionProxyGroup 表示策略组段。
	sectionProxyGroup shadowrocketSection = "[Proxy Group]"
	// sectionRule 表示分流规则段。
	sectionRule shadowrocketSection = "[Rule]"
)

type proxyProtocol string

const (
	// proxyProtocolSS 表示 Shadowrocket Shadowsocks 协议关键字。
	proxyProtocolSS proxyProtocol = "ss"
	// proxyProtocolSSR 表示 Shadowrocket ShadowsocksR 协议关键字。
	proxyProtocolSSR proxyProtocol = "ssr"
	// proxyProtocolTrojan 表示 Shadowrocket Trojan 协议关键字。
	proxyProtocolTrojan proxyProtocol = "trojan"
	// proxyProtocolVMess 表示 Shadowrocket VMess 协议关键字。
	proxyProtocolVMess proxyProtocol = "vmess"
	// proxyProtocolVLESS 表示 Shadowrocket VLESS 协议关键字。
	proxyProtocolVLESS proxyProtocol = "vless"
	// proxyProtocolHysteria2 表示 Shadowrocket Hysteria2 协议关键字。
	proxyProtocolHysteria2 proxyProtocol = "hysteria2"
	// proxyProtocolTUIC 表示 Shadowrocket TUIC 协议关键字。
	proxyProtocolTUIC proxyProtocol = "tuic"
)

type groupType string

const (
	// groupTypeSelect 表示手动选择策略组。
	groupTypeSelect groupType = "select"
	// groupTypeURLTest 表示按延迟测试自动选择策略组。
	groupTypeURLTest groupType = "url-test"
)

type ruleType string

const (
	// ruleTypeDomain 表示域名精确匹配规则。
	ruleTypeDomain ruleType = "DOMAIN"
	// ruleTypeDomainSuffix 表示域名后缀匹配规则。
	ruleTypeDomainSuffix ruleType = "DOMAIN-SUFFIX"
	// ruleTypeDomainKeyword 表示域名关键字匹配规则。
	ruleTypeDomainKeyword ruleType = "DOMAIN-KEYWORD"
	// ruleTypeDomainRegex 表示域名正则匹配规则。
	ruleTypeDomainRegex ruleType = "DOMAIN-REGEX"
	// ruleTypeIPCIDR 表示 IPv4 CIDR 匹配规则。
	ruleTypeIPCIDR ruleType = "IP-CIDR"
	// ruleTypeIPCIDR6 表示 IPv6 CIDR 匹配规则。
	ruleTypeIPCIDR6 ruleType = "IP-CIDR6"
	// ruleTypeGEOIP 表示国家或地区 IP 库匹配规则。
	ruleTypeGEOIP ruleType = "GEOIP"
	// ruleTypeRuleSet 表示远程规则集引用。
	ruleTypeRuleSet ruleType = "RULE-SET"
	// ruleTypeFinal 表示兜底规则。
	ruleTypeFinal ruleType = "FINAL"
)

type parameterKey string

const (
	// parameterEncryptMethod 表示加密方法参数。
	parameterEncryptMethod parameterKey = "encrypt-method"
	// parameterPassword 表示密码参数。
	parameterPassword parameterKey = "password"
	// parameterUsername 表示用户标识参数，VMess/VLESS 通常填写 UUID。
	parameterUsername parameterKey = "username"
	// parameterUUID 表示 TUIC UUID 参数。
	parameterUUID parameterKey = "uuid"
	// parameterUDPRelay 表示是否启用 UDP 转发。
	parameterUDPRelay parameterKey = "udp-relay"
	// parameterTLS 表示是否启用 TLS。
	parameterTLS parameterKey = "tls"
	// parameterSNI 表示 TLS SNI。
	parameterSNI parameterKey = "sni"
	// parameterSkipCertVerify 表示是否跳过证书校验。
	parameterSkipCertVerify parameterKey = "skip-cert-verify"
	// parameterALPN 表示 TLS ALPN 列表。
	parameterALPN parameterKey = "alpn"
	// parameterFingerprint 表示 uTLS 指纹伪装。
	parameterFingerprint parameterKey = "fingerprint"
	// parameterReality 表示是否启用 Reality。
	parameterReality parameterKey = "reality"
	// parameterRealityPublicKey 表示 Reality 服务器公钥。
	parameterRealityPublicKey parameterKey = "public-key"
	// parameterRealityShortID 表示 Reality short id。
	parameterRealityShortID parameterKey = "short-id"
	// parameterObfs 表示混淆类型。
	parameterObfs parameterKey = "obfs"
	// parameterObfsHost 表示 Shadowsocks 插件混淆 Host。
	parameterObfsHost parameterKey = "obfs-host"
	// parameterObfsURI 表示 Shadowsocks 插件混淆路径。
	parameterObfsURI parameterKey = "obfs-uri"
	// parameterObfsParam 表示 SSR 混淆参数。
	parameterObfsParam parameterKey = "obfs-param"
	// parameterObfsPassword 表示 Hysteria2 混淆密码。
	parameterObfsPassword parameterKey = "obfs-password"
	// parameterProtocol 表示 SSR 协议参数。
	parameterProtocol parameterKey = "protocol"
	// parameterProtocolParam 表示 SSR 协议附加参数。
	parameterProtocolParam parameterKey = "protocol-param"
	// parameterFlow 表示 VLESS flow 参数。
	parameterFlow parameterKey = "flow"
	// parameterWS 表示是否启用 WebSocket 传输。
	parameterWS parameterKey = "ws"
	// parameterWSPath 表示 WebSocket 路径。
	parameterWSPath parameterKey = "ws-path"
	// parameterWSHeaders 表示 WebSocket 请求头。
	parameterWSHeaders parameterKey = "ws-headers"
	// parameterGRPC 表示是否启用 gRPC 传输。
	parameterGRPC parameterKey = "grpc"
	// parameterGRPCServiceName 表示 gRPC service name。
	parameterGRPCServiceName parameterKey = "grpc-service-name"
	// parameterHTTPUpgrade 表示是否启用 HTTP Upgrade 传输。
	parameterHTTPUpgrade parameterKey = "http-upgrade"
	// parameterHTTPUpgradePath 表示 HTTP Upgrade 路径。
	parameterHTTPUpgradePath parameterKey = "http-upgrade-path"
	// parameterHTTPUpgradeHost 表示 HTTP Upgrade Host。
	parameterHTTPUpgradeHost parameterKey = "http-upgrade-host"
	// parameterUp 表示 Hysteria2 上行带宽。
	parameterUp parameterKey = "up"
	// parameterDown 表示 Hysteria2 下行带宽。
	parameterDown parameterKey = "down"
	// parameterCongestionControl 表示 TUIC 拥塞控制算法。
	parameterCongestionControl parameterKey = "congestion-controller"
	// parameterUDPRelayMode 表示 TUIC UDP 转发模式。
	parameterUDPRelayMode parameterKey = "udp-relay-mode"
	// parameterZeroRTTHandshake 表示 TUIC 是否启用 0-RTT 握手。
	parameterZeroRTTHandshake parameterKey = "zero-rtt-handshake"
	// parameterURL 表示 url-test 策略组探测地址。
	parameterURL parameterKey = "url"
	// parameterInterval 表示 url-test 策略组探测间隔。
	parameterInterval parameterKey = "interval"
	// parameterTolerance 表示 url-test 策略组延迟容忍值。
	parameterTolerance parameterKey = "tolerance"
)

type transportType string

const (
	// transportTypeWebSocket 表示 WebSocket 传输。
	transportTypeWebSocket transportType = "ws"
	// transportTypeGRPC 表示 gRPC 传输。
	transportTypeGRPC transportType = "grpc"
	// transportTypeHTTPUpgrade 表示 HTTP Upgrade 传输。
	transportTypeHTTPUpgrade transportType = "httpupgrade"
)

type renderContext struct {
	// proxyNames 记录 sing-box outbound tag 到 Shadowrocket 节点名称的映射。
	proxyNames map[string]string
	// proxyTags 按成功导出的顺序保存 outbound tag，保证策略组成员顺序稳定。
	proxyTags []string
	// groupNames 记录节点分组 tag 到 Shadowrocket 策略组名称的映射。
	groupNames map[string]string
	// warnings 收集可降级问题，最后统一写入日志，避免中断整份配置生成。
	warnings []string
}

// Render 将统一中间数据渲染为 Shadowrocket INI 风格配置文本。
// 函数不访问数据库、不发 HTTP 请求，只处理传入的 DNS、Outbound、节点分组和规则集，
// 因此协议映射、分组一致性和规则展开都可以通过单元测试直接覆盖。
func Render(deviceCode string, dns entity.SingDNS, outbounds []entity.SingBoxOut, groups []*entity.NodeGroup, ruleSets []*entity.RuleSet) string {
	ctx := &renderContext{
		proxyNames: make(map[string]string),
		groupNames: make(map[string]string),
	}

	proxyLines := renderProxySection(ctx, outbounds)
	groupLines := renderProxyGroupSection(ctx, deviceCode, groups)
	ruleLines := renderRuleSection(ctx, deviceCode, ruleSets)
	ctx.flushWarnings()

	sections := [][]string{
		renderGeneralSection(dns),
		withSectionHeader(sectionProxy, proxyLines),
		withSectionHeader(sectionProxyGroup, groupLines),
		withSectionHeader(sectionRule, ruleLines),
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
		string(sectionGeneral),
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
			ctx.warnf("Shadowrocket proxy name duplicated, skip outbound: tag=%s name=%s", outbound.Tag, name)
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
	case entity.OutboundProtocolShadowsocksR:
		return renderShadowsocksRLine(ctx, outbound)
	case entity.OutboundProtocolTrojan:
		return renderTrojanLine(ctx, outbound)
	case entity.OutboundProtocolVMess:
		return renderVMessLine(ctx, outbound)
	case entity.OutboundProtocolVLESS:
		return renderVLESSLine(ctx, outbound)
	case entity.OutboundProtocolHysteria2:
		return renderHysteria2Line(ctx, outbound)
	case entity.OutboundProtocolTUIC:
		return renderTUICLine(ctx, outbound)
	case entity.OutboundProtocolHysteria:
		ctx.warnf("Shadowrocket hysteria v1 outbound is not exported in this version, skip: tag=%s", outbound.Tag)
		return "", "", false
	default:
		if outbound.Type != "" && outbound.Type != string(entity.NodeGroupTypeSelector) && outbound.Type != string(entity.NodeGroupTypeURLTest) && outbound.Type != string(entity.OutboundProtocolDirect) {
			ctx.warnf("Shadowrocket unrecognized outbound protocol, skip: tag=%s type=%s", outbound.Tag, outbound.Type)
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
		ctx.warnf("Shadowrocket shadowsocks outbound missing method or password, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	parts := []string{
		name + " = " + string(proxyProtocolSS),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue(parameterEncryptMethod, outbound.Method),
		keyValue(parameterPassword, outbound.Password),
		keyValue(parameterUDPRelay, "true"),
	}
	appendShadowsocksPluginParameters(&parts, outbound)
	return strings.Join(parts, ", "), name, true
}

func renderShadowsocksRLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	if outbound.Method == "" || outbound.Password == "" || outbound.Protocol == "" {
		ctx.warnf("Shadowrocket shadowsocksr outbound missing method, password or protocol, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	parts := []string{
		name + " = " + string(proxyProtocolSSR),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue(parameterEncryptMethod, outbound.Method),
		keyValue(parameterPassword, outbound.Password),
		keyValue(parameterProtocol, outbound.Protocol),
		keyValue(parameterUDPRelay, "true"),
	}
	if outbound.ProtocolParam != "" {
		parts = append(parts, keyValue(parameterProtocolParam, outbound.ProtocolParam))
	}
	if outbound.Obfs != nil && outbound.Obfs.Value != "" {
		parts = append(parts, keyValue(parameterObfs, outbound.Obfs.Value))
	}
	if outbound.ObfsParam != "" {
		parts = append(parts, keyValue(parameterObfsParam, outbound.ObfsParam))
	}
	return strings.Join(parts, ", "), name, true
}

func renderTrojanLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	if outbound.Password == "" {
		ctx.warnf("Shadowrocket trojan outbound missing password, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	parts := []string{
		name + " = " + string(proxyProtocolTrojan),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue(parameterPassword, outbound.Password),
		keyValue(parameterUDPRelay, "true"),
	}
	appendTLSParameters(&parts, outbound.TLS)
	appendTransportParameters(&parts, outbound.Transport)
	return strings.Join(parts, ", "), name, true
}

func renderVMessLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	if outbound.UUID == "" {
		ctx.warnf("Shadowrocket vmess outbound missing uuid, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	parts := []string{
		name + " = " + string(proxyProtocolVMess),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue(parameterUsername, outbound.UUID),
		keyValue(parameterUDPRelay, "true"),
	}
	if outbound.Security != "" && outbound.Security != "auto" {
		parts = append(parts, keyValue(parameterEncryptMethod, outbound.Security))
	}
	appendTLSParameters(&parts, outbound.TLS)
	appendTransportParameters(&parts, outbound.Transport)
	return strings.Join(parts, ", "), name, true
}

func renderVLESSLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	if outbound.UUID == "" {
		ctx.warnf("Shadowrocket vless outbound missing uuid, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	parts := []string{
		name + " = " + string(proxyProtocolVLESS),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue(parameterUsername, outbound.UUID),
		keyValue(parameterUDPRelay, "true"),
	}
	if outbound.Flow != "" {
		parts = append(parts, keyValue(parameterFlow, outbound.Flow))
	}
	appendTLSParameters(&parts, outbound.TLS)
	appendTransportParameters(&parts, outbound.Transport)
	return strings.Join(parts, ", "), name, true
}

func renderHysteria2Line(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	if outbound.Password == "" {
		ctx.warnf("Shadowrocket hysteria2 outbound missing password, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	parts := []string{
		name + " = " + string(proxyProtocolHysteria2),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue(parameterPassword, outbound.Password),
	}
	appendTLSParameters(&parts, outbound.TLS)
	appendBandwidthParameters(&parts, outbound)
	appendHysteria2ObfsParameters(&parts, outbound.Obfs)
	return strings.Join(parts, ", "), name, true
}

func renderTUICLine(ctx *renderContext, outbound entity.SingBoxOut) (string, string, bool) {
	name, ok := validateProxyBasics(ctx, outbound)
	if !ok {
		return "", "", false
	}
	if outbound.UUID == "" || outbound.Password == "" {
		ctx.warnf("Shadowrocket tuic outbound missing uuid or password, skip: tag=%s", outbound.Tag)
		return "", "", false
	}
	parts := []string{
		name + " = " + string(proxyProtocolTUIC),
		outbound.Server,
		strconv.Itoa(outbound.ServerPort),
		keyValue(parameterUUID, outbound.UUID),
		keyValue(parameterPassword, outbound.Password),
	}
	appendTLSParameters(&parts, outbound.TLS)
	if outbound.CongestionController != "" {
		parts = append(parts, keyValue(parameterCongestionControl, outbound.CongestionController))
	}
	if outbound.UdpRelayMode != "" {
		parts = append(parts, keyValue(parameterUDPRelayMode, outbound.UdpRelayMode))
	}
	if outbound.ZeroRttHandshake {
		parts = append(parts, keyValue(parameterZeroRTTHandshake, "true"))
	}
	return strings.Join(parts, ", "), name, true
}

func validateProxyBasics(ctx *renderContext, outbound entity.SingBoxOut) (string, bool) {
	name := policyName(outbound.Tag)
	if name == "" {
		ctx.warnf("Shadowrocket outbound missing tag, skip: type=%s server=%s", outbound.Type, outbound.Server)
		return "", false
	}
	if strings.TrimSpace(outbound.Server) == "" || outbound.ServerPort <= 0 {
		ctx.warnf("Shadowrocket outbound missing server or server_port, skip: tag=%s", outbound.Tag)
		return "", false
	}
	return name, true
}

func appendTLSParameters(parts *[]string, tls *entity.SingTLS) {
	if tls == nil || !tls.Enabled {
		return
	}
	*parts = append(*parts, keyValue(parameterTLS, "true"))
	if tls.ServerName != "" {
		*parts = append(*parts, keyValue(parameterSNI, tls.ServerName))
	}
	if tls.Insecure {
		*parts = append(*parts, keyValue(parameterSkipCertVerify, "true"))
	}
	if len(tls.Alpn) > 0 {
		*parts = append(*parts, keyValue(parameterALPN, strings.Join(tls.Alpn, "|")))
	}
	if tls.Utls != nil && tls.Utls.Enabled && tls.Utls.Fingerprint != "" {
		*parts = append(*parts, keyValue(parameterFingerprint, tls.Utls.Fingerprint))
	}
	if tls.Reality != nil && tls.Reality.Enabled {
		*parts = append(*parts, keyValue(parameterReality, "true"))
		if tls.Reality.PublicKey != "" {
			*parts = append(*parts, keyValue(parameterRealityPublicKey, tls.Reality.PublicKey))
		}
		if tls.Reality.ShortID != "" {
			*parts = append(*parts, keyValue(parameterRealityShortID, tls.Reality.ShortID))
		}
	}
}

func appendShadowsocksPluginParameters(parts *[]string, outbound entity.SingBoxOut) {
	pluginOptions := parseSemicolonOptions(outbound.PluginOpts)
	if obfs := firstNonEmpty(pluginOptions[string(parameterObfs)], outbound.Plugin); obfs == "http" || obfs == "tls" {
		*parts = append(*parts, keyValue(parameterObfs, obfs))
	}
	if obfsHost := firstNonEmpty(pluginOptions[string(parameterObfsHost)], outbound.ObfsParam); obfsHost != "" {
		*parts = append(*parts, keyValue(parameterObfsHost, obfsHost))
	}
	if obfsURI := pluginOptions[string(parameterObfsURI)]; obfsURI != "" {
		*parts = append(*parts, keyValue(parameterObfsURI, obfsURI))
	}
}

func appendTransportParameters(parts *[]string, transport *entity.SingTransport) {
	if transport == nil {
		return
	}
	switch transportType(transport.Type) {
	case transportTypeWebSocket:
		*parts = append(*parts, keyValue(parameterWS, "true"))
		if transport.Path != "" {
			*parts = append(*parts, keyValue(parameterWSPath, transport.Path))
		}
		if host := firstHeaderValue(transport.Headers, "Host"); host != "" {
			*parts = append(*parts, keyValue(parameterWSHeaders, "Host:"+host))
		}
	case transportTypeGRPC:
		*parts = append(*parts, keyValue(parameterGRPC, "true"))
		if transport.ServiceName != "" {
			*parts = append(*parts, keyValue(parameterGRPCServiceName, transport.ServiceName))
		}
	case transportTypeHTTPUpgrade:
		*parts = append(*parts, keyValue(parameterHTTPUpgrade, "true"))
		if transport.Path != "" {
			*parts = append(*parts, keyValue(parameterHTTPUpgradePath, transport.Path))
		}
		if host, ok := transport.Host.(string); ok && host != "" {
			*parts = append(*parts, keyValue(parameterHTTPUpgradeHost, host))
		}
	}
}

func appendBandwidthParameters(parts *[]string, outbound entity.SingBoxOut) {
	if outbound.Up != "" {
		*parts = append(*parts, keyValue(parameterUp, outbound.Up))
	} else if outbound.UpMbps > 0 {
		*parts = append(*parts, keyValue(parameterUp, strconv.Itoa(outbound.UpMbps)+" Mbps"))
	}
	if outbound.Down != "" {
		*parts = append(*parts, keyValue(parameterDown, outbound.Down))
	} else if outbound.DownMbps > 0 {
		*parts = append(*parts, keyValue(parameterDown, strconv.Itoa(outbound.DownMbps)+" Mbps"))
	}
}

func appendHysteria2ObfsParameters(parts *[]string, obfs *entity.SingObfs) {
	if obfs == nil || obfs.Value == "" {
		return
	}
	if obfs.Type != "" {
		*parts = append(*parts, keyValue(parameterObfs, obfs.Type))
	}
	*parts = append(*parts, keyValue(parameterObfsPassword, obfs.Value))
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
		if len(members) == 0 {
			continue
		}
		groupName := policyName(group.Tag)
		ctx.groupNames[group.Tag] = groupName

		// 按设备解析最终分组类型，确保 url-test 关键字与 url/interval/tolerance 参数基于同一次判定，
		// 避免设备覆盖为 urltest 时类型变了但探测参数缺失。
		groupType := common.ResolveGroupType(group, deviceCode)

		lineMembers := make([]string, 0, len(members)+4)
		lineMembers = append(lineMembers, groupName+" = "+string(resolveGroupType(groupType)))
		for _, tag := range members {
			lineMembers = append(lineMembers, ctx.proxyNames[tag])
		}
		if groupType == entity.NodeGroupTypeURLTest {
			testURL := strings.TrimSpace(group.TestURL)
			if testURL == "" {
				testURL = defaultURLTestURL
			}
			lineMembers = append(lineMembers, keyValue(parameterURL, testURL), keyValue(parameterInterval, "600"), keyValue(parameterTolerance, "50"))
		}
		lines = append(lines, strings.Join(lineMembers, ", "))
	}
	return lines
}

func exportedProxyTags(ctx *renderContext) []string {
	return append([]string(nil), ctx.proxyTags...)
}

func resolveGroupType(resolved entity.NodeGroupType) groupType {
	if resolved == entity.NodeGroupTypeURLTest {
		return groupTypeURLTest
	}
	return groupTypeSelect
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
		// 规则引用的出站/策略组在当前设备配置中不存在时跳过该条规则，避免生成悬空策略引用。
		policy, ok := ctx.resolvePolicy(ruleSet.Outbound)
		if !ok {
			ctx.warnf("Shadowrocket ruleset outbound not found in proxies or groups, skip rule: tag=%s outbound=%s", ruleSet.Tag, ruleSet.Outbound)
			continue
		}
		if entity.RuleSetType(ruleSet.RuleSetType) == entity.RuleSetTypeRemote {
			if strings.TrimSpace(ruleSet.URL) == "" {
				ctx.warnf("Shadowrocket remote ruleset missing url, skip: tag=%s", ruleSet.Tag)
				continue
			}
			lines = append(lines, strings.Join([]string{string(ruleTypeRuleSet), ruleSet.URL, policy}, ","))
			continue
		}
		lines = append(lines, expandLocalRuleSet(ctx, ruleSet, policy)...)
	}
	lines = append(lines, strings.Join([]string{string(ruleTypeFinal), ctx.policyReference(defaultPolicyName)}, ","))
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
		ctx.warnf("Shadowrocket local ruleset content invalid, skip: tag=%s err=%v", ruleSet.Tag, err)
		return nil
	}
	lines := make([]string, 0, len(rules))
	for _, rule := range rules {
		lines = appendRules(lines, ruleTypeDomain, rule.Domain, policy)
		lines = appendRules(lines, ruleTypeDomainSuffix, rule.DomainSuffix, policy)
		lines = appendRules(lines, ruleTypeDomainKeyword, rule.DomainKeyword, policy)
		lines = appendRules(lines, ruleTypeDomainRegex, rule.DomainRegex, policy)
		for _, cidr := range rule.IPCIDR {
			itemType := ruleTypeIPCIDR
			if ip, _, err := net.ParseCIDR(cidr); err == nil && ip.To4() == nil {
				itemType = ruleTypeIPCIDR6
			}
			lines = append(lines, strings.Join([]string{string(itemType), cidr, policy}, ","))
		}
		for _, geoip := range rule.GEOIP {
			lines = append(lines, strings.Join([]string{string(ruleTypeGEOIP), strings.ToUpper(geoip), policy}, ","))
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

func appendRules(lines []string, itemType ruleType, values []string, policy string) []string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		lines = append(lines, strings.Join([]string{string(itemType), value, policy}, ","))
	}
	return lines
}

func (ctx *renderContext) policyReference(tag string) string {
	name, _ := ctx.resolvePolicy(tag)
	return name
}

// resolvePolicy 把规则出站标签解析为 Shadowrocket 策略名称，并返回该策略是否真实存在。
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

func keyValue(key parameterKey, value string) string {
	return string(key) + "=" + encodeValue(value)
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

func withSectionHeader(header shadowrocketSection, lines []string) []string {
	if len(lines) == 0 {
		return []string{string(header)}
	}
	result := make([]string, 0, len(lines)+1)
	result = append(result, string(header))
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
