package singbox

import (
	"encoding/json"
	"singboxconfig/entity"

	"github.com/sirupsen/logrus"
)

// GetDefaultDNS 返回与历史硬编码一致的默认 DNS 配置。
// 当存储里没有 dns_config，或保存值非法时，生成链路会回退到这里。
func GetDefaultDNS() entity.SingDNS {
	dnsServers := []entity.SingDNSServer{
		{
			Tag:             "dns_proxy",
			Address:         "https://1.1.1.1/dns-query",
			AddressResolver: "dns_resolver",
			Strategy:        "ipv4_only",
			Detour:          "general",
		},
		{
			Tag:             "dns_direct",
			Address:         "h3://dns.alidns.com/dns-query",
			AddressResolver: "dns_resolver",
			Strategy:        "ipv4_only",
			Detour:          "direct",
		},
		{
			Tag:     "dns_block",
			Address: "rcode://refused",
		},
		{
			Tag:      "dns_resolver",
			Address:  "223.5.5.5",
			Strategy: "ipv4_only",
			Detour:   "direct",
		},
	}

	dnsRules := []entity.SingDNSRule{
		{
			Outbound: "any",
			Server:   "dns_resolver",
		},
		{
			ClashMode: "direct",
			Server:    "dns_direct",
		},
		{
			ClashMode: "global",
			Server:    "dns_proxy",
		},
		{
			RuleSet: "cnip",
			Server:  "dns_direct",
		},
		{
			RuleSet: "cnsite",
			Server:  "dns_direct",
		},
	}

	return entity.SingDNS{
		Servers: dnsServers,
		Rules:   dnsRules,
		Final:   "dns_proxy",
	}
}

// ResolveDNS 按存储值解析 DNS 配置。
// 只要 JSON 为空或不合法，就统一回退到默认 DNS，避免生成接口因配置损坏直接失败。
func ResolveDNS(configJSON string) entity.SingDNS {
	if configJSON == "" {
		return GetDefaultDNS()
	}

	var dns entity.SingDNS
	if err := json.Unmarshal([]byte(configJSON), &dns); err != nil {
		logrus.WithError(err).Warn("ResolveDNS: invalid dns_config, fallback to default")
		return GetDefaultDNS()
	}

	return dns
}
