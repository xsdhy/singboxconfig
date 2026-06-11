package surge

import (
	"singboxconfig/entity"
	"strings"
	"testing"
)

func TestRenderProxyProtocols(t *testing.T) {
	cfg := Render("phone", "", "", []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "ss-node",
			Server:     "ss.example.com",
			ServerPort: 8388,
			Method:     "chacha20-ietf-poly1305",
			Password:   "ss-pass",
			Plugin:     "obfs-local",
			PluginOpts: "obfs=tls;obfs-host=cdn.example.com;obfs-uri=/",
		},
		{
			Type:       string(entity.OutboundProtocolTrojan),
			Tag:        "trojan-node",
			Server:     "trojan.example.com",
			ServerPort: 443,
			Password:   "trojan-pass",
			TLS: &entity.SingTLS{
				Enabled:    true,
				ServerName: "sni.example.com",
				Insecure:   true,
			},
		},
		{
			Type:       string(entity.OutboundProtocolVMess),
			Tag:        "vmess-node",
			Server:     "vmess.example.com",
			ServerPort: 443,
			UUID:       "0233d11c-15a4-47d3-ade3-48ffca0ce119",
			Security:   "aes-128-gcm",
			TLS:        &entity.SingTLS{Enabled: true, ServerName: "vmess-sni.example.com"},
			Transport: &entity.SingTransport{
				Type:    "ws",
				Path:    "/ws",
				Headers: map[string][]string{"Host": {"host.example.com"}},
			},
		},
		{
			Type:       string(entity.OutboundProtocolHTTP),
			Tag:        "http-node",
			Server:     "http.example.com",
			ServerPort: 8080,
			Username:   "http-user",
			Password:   "http-pass",
		},
		{
			Type:       string(entity.OutboundProtocolHTTP),
			Tag:        "https-node",
			Server:     "https.example.com",
			ServerPort: 443,
			Username:   "https-user",
			Password:   "https-pass",
			TLS: &entity.SingTLS{
				Enabled:    true,
				ServerName: "https-sni.example.com",
			},
		},
	}, nil, nil, nil, nil)

	mustContain(t, cfg, "ss-node = ss, ss.example.com, 8388, encrypt-method=chacha20-ietf-poly1305, password=ss-pass, udp-relay=true, obfs=tls, obfs-host=cdn.example.com, obfs-uri=/")
	mustContain(t, cfg, "trojan-node = trojan, trojan.example.com, 443, password=trojan-pass, udp-relay=true, tls=true, sni=sni.example.com, skip-cert-verify=true")
	mustContain(t, cfg, "vmess-node = vmess, vmess.example.com, 443, username=0233d11c-15a4-47d3-ade3-48ffca0ce119, encrypt-method=aes-128-gcm, tls=true, sni=vmess-sni.example.com, ws=true, ws-path=/ws, ws-headers=Host:host.example.com")
	mustContain(t, cfg, "http-node = http, http.example.com, 8080, username=http-user, password=http-pass")
	mustContain(t, cfg, "https-node = https, https.example.com, 443, username=https-user, password=https-pass, tls=true, sni=https-sni.example.com")
}

func TestRenderWireGuardEndpoints(t *testing.T) {
	cfg := Render("phone", "", "", []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "ss-node",
			Server:     "ss.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "pass",
		},
	}, []entity.SingEndpointWireguard{
		{
			Type:       "wireguard",
			Tag:        "wg-node",
			MTU:        1280,
			Address:    []string{"10.0.2.2/32", "fd00:1111::11/128"},
			PrivateKey: "client-private-key",
			Peers: []entity.SingEndpointWireguardPeer{
				{
					Address:                     "wg.example.com",
					Port:                        51820,
					PublicKey:                   "peer-public-key",
					PreSharedKey:                "peer-psk",
					AllowedIps:                  []string{"0.0.0.0/0", "::/0"},
					PersistentKeepaliveInterval: 25,
				},
			},
		},
	}, nil, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector)},
	}, nil)

	mustContain(t, cfg, "wg-node = wireguard, section-name=wg-node")
	mustContain(t, cfg, "[WireGuard wg-node]")
	mustContain(t, cfg, "private-key = client-private-key")
	mustContain(t, cfg, "self-ip = 10.0.2.2\n")
	mustContain(t, cfg, "self-ip-v6 = fd00:1111::11\n")
	mustContain(t, cfg, "mtu = 1280")
	mustContain(t, cfg, "peer = (public-key = peer-public-key, allowed-ips = \"0.0.0.0/0, ::/0\", endpoint = wg.example.com:51820, preshared-key = peer-psk, keepalive = 25)")
	mustContain(t, cfg, "general = select, ss-node, wg-node")
}

func TestRenderSkipsUnsupportedProtocolAndKeepsGroupReferencesValid(t *testing.T) {
	cfg := Render("phone", "", "", []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "香港01",
			Server:     "hk.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "pass",
		},
		{
			Type:       string(entity.OutboundProtocolVLESS),
			Tag:        "香港VLESS",
			Server:     "vless.example.com",
			ServerPort: 443,
			UUID:       "00000000-0000-0000-0000-000000000001",
		},
	}, nil, nil, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector), Include: "香港"},
	}, nil)

	mustContain(t, cfg, "香港01 = ss, hk.example.com, 8388")
	mustNotContain(t, cfg, "香港VLESS =")
	mustContain(t, cfg, "general = select, 香港01")
	mustNotContain(t, cfg, "general = select, 香港01, 香港VLESS")
}

func TestRenderExpandsRuleSets(t *testing.T) {
	ruleContent := `{"rules":[{"domain":["exact.example.com"],"domain_suffix":["example.com"],"domain_keyword":["video"],"domain_regex":["^api\\."],"ip_cidr":["192.168.0.0/16","2001:db8::/32"],"geoip":["cn"],"process_name":["WeChat"],"process_path":["/Applications/QoderWork CN.app/Contents/MacOS/QoderWork CN"]}]}`
	cfg := Render("phone", "", "", []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "proxy-node",
			Server:     "proxy.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "pass",
		},
	}, nil, nil, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector)},
	}, []*entity.RuleSet{
		{Tag: "remote", RuleSetType: string(entity.RuleSetTypeRemote), URL: "https://example.com/remote.list", Outbound: "general", Sort: 10},
		{Tag: "local", RuleSetType: string(entity.RuleSetTypeLocal), Content: ruleContent, Outbound: "general", Sort: 20},
		{Tag: "bad", RuleSetType: string(entity.RuleSetTypeLocal), Content: `{`, Outbound: "general", Sort: 30},
	})

	mustContain(t, cfg, "RULE-SET,https://example.com/remote.list,general")
	mustContain(t, cfg, "DOMAIN,exact.example.com,general")
	mustContain(t, cfg, "DOMAIN-SUFFIX,example.com,general")
	mustContain(t, cfg, "DOMAIN-KEYWORD,video,general")
	mustContain(t, cfg, "DOMAIN-REGEX,^api\\.,general")
	mustContain(t, cfg, "IP-CIDR,192.168.0.0/16,general")
	mustContain(t, cfg, "IP-CIDR6,2001:db8::/32,general")
	mustContain(t, cfg, "GEOIP,CN,general,no-resolve")
	mustContain(t, cfg, "PROCESS-NAME,WeChat,general")
	mustContain(t, cfg, "PROCESS-NAME,\"/Applications/QoderWork CN.app/Contents/MacOS/QoderWork CN\",general")
	mustContain(t, cfg, "FINAL,general")
}

// TestRenderProxyGroupDeviceTypeOverride 验证 Surge 分组类型可按设备覆盖。
// 同一份分组定义，对 gateway 覆盖为 urltest（输出 url-test 并携带探测参数），
// phone 未命中覆盖时回退默认 selector（输出 select 且不带探测参数）。
func TestRenderProxyGroupDeviceTypeOverride(t *testing.T) {
	outbounds := []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "香港01",
			Server:     "hk.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "pass",
		},
	}
	groups := []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector), Include: "香港", DeviceTypeOverrides: "gateway:urltest"},
	}

	gatewayCfg := Render("gateway", "", "", outbounds, nil, nil, groups, nil)
	mustContain(t, gatewayCfg, "general = url-test, 香港01, url=https://www.gstatic.com/generate_204, interval=600, tolerance=50")

	phoneCfg := Render("phone", "", "", outbounds, nil, nil, groups, nil)
	mustContain(t, phoneCfg, "general = select, 香港01\n")
	mustNotContain(t, phoneCfg, "url-test")
}

// TestRenderSkipsRuleWithUnknownOutbound 验证规则引用的出站/策略组不存在时跳过该条规则，
// 已存在出站、内置 DIRECT、以及 FINAL 兜底不受影响。
func TestRenderSkipsRuleWithUnknownOutbound(t *testing.T) {
	ruleContent := `{"rules":[{"domain_suffix":["example.com"]}]}`
	cfg := Render("phone", "", "", []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "proxy-node",
			Server:     "proxy.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "pass",
		},
	}, nil, nil, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector)},
	}, []*entity.RuleSet{
		{Tag: "ok-group", RuleSetType: string(entity.RuleSetTypeLocal), Content: ruleContent, Outbound: "general", Sort: 10},
		{Tag: "ok-proxy", RuleSetType: string(entity.RuleSetTypeLocal), Content: ruleContent, Outbound: "proxy-node", Sort: 20},
		{Tag: "ok-direct", RuleSetType: string(entity.RuleSetTypeLocal), Content: ruleContent, Outbound: "direct", Sort: 30},
		{Tag: "ghost", RuleSetType: string(entity.RuleSetTypeLocal), Content: ruleContent, Outbound: "missing-node", Sort: 40},
		{Tag: "ghost-remote", RuleSetType: string(entity.RuleSetTypeRemote), URL: "https://example.com/ghost.list", Outbound: "missing-node", Sort: 50},
	})

	mustContain(t, cfg, "DOMAIN-SUFFIX,example.com,general")
	mustContain(t, cfg, "DOMAIN-SUFFIX,example.com,proxy-node")
	mustContain(t, cfg, "DOMAIN-SUFFIX,example.com,DIRECT")
	mustNotContain(t, cfg, "missing-node")
	mustNotContain(t, cfg, "https://example.com/ghost.list")
	mustContain(t, cfg, "FINAL,general")
}

func mustContain(t *testing.T, text string, want string) {
	t.Helper()
	if !strings.Contains(text, want) {
		t.Fatalf("rendered config missing %q:\n%s", want, text)
	}
}

func mustNotContain(t *testing.T, text string, unexpected string) {
	t.Helper()
	if strings.Contains(text, unexpected) {
		t.Fatalf("rendered config contains unexpected %q:\n%s", unexpected, text)
	}
}

// TestRenderLocalRuleSetURLReference 验证配置系统 Host 后，规则条数达到阈值的 local 规则集输出为单行
// RULE-SET,<url>,<policy>，不再逐条展开，且 URL 指向本服务 open 接口并正确转义（software / device 走 query）。
func TestRenderLocalRuleSetURLReference(t *testing.T) {
	cfg := Render("iphone 15", "tok en", "https://config.example.com/", []entity.SingBoxOut{
		{Tag: "香港01", Type: string(entity.OutboundProtocolShadowsocks), Server: "hk.example.com", ServerPort: 8388, Method: "aes-128-gcm", Password: "pass"},
	}, nil, nil, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector), Include: "香港"},
	}, []*entity.RuleSet{
		{Tag: "geosite-cn", RuleSetType: string(entity.RuleSetTypeLocal), Outbound: "general", Content: `{"version":1,"rules":[{"domain_suffix":["cn","com","net"]}]}`},
	})

	wantLine := "RULE-SET,https://config.example.com/open/rules/geosite-cn?device=iphone+15&software=surge&token=tok+en,general"
	if !strings.Contains(cfg, wantLine) {
		t.Errorf("config missing expected rule-set line:\n%s\n---\n%s", wantLine, cfg)
	}
	if strings.Contains(cfg, "DOMAIN-SUFFIX,cn") {
		t.Errorf("local ruleset should not be expanded when system host configured:\n%s", cfg)
	}
}

// TestRenderLocalRuleSetInlineBelowThreshold 验证即便配置了系统 Host，规则条数少于阈值的 local 规则集
// 仍逐条内联展开，而不是输出 RULE-SET URL 引用。
func TestRenderLocalRuleSetInlineBelowThreshold(t *testing.T) {
	cfg := Render("iphone 15", "tok en", "https://config.example.com/", []entity.SingBoxOut{
		{Tag: "香港01", Type: string(entity.OutboundProtocolShadowsocks), Server: "hk.example.com", ServerPort: 8388, Method: "aes-128-gcm", Password: "pass"},
	}, nil, nil, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector), Include: "香港"},
	}, []*entity.RuleSet{
		{Tag: "geosite-cn", RuleSetType: string(entity.RuleSetTypeLocal), Outbound: "general", Content: `{"version":1,"rules":[{"domain_suffix":["cn","com"]}]}`},
	})

	if !strings.Contains(cfg, "DOMAIN-SUFFIX,cn,general") || !strings.Contains(cfg, "DOMAIN-SUFFIX,com,general") {
		t.Errorf("ruleset below threshold should be expanded inline:\n%s", cfg)
	}
	if strings.Contains(cfg, "/open/rules/geosite-cn") {
		t.Errorf("ruleset below threshold should not emit a RULE-SET url reference:\n%s", cfg)
	}
}

// TestRenderLocalRuleSetFallbackExpand 验证未配置系统 Host 时，local 规则集仍逐条展开。
func TestRenderLocalRuleSetFallbackExpand(t *testing.T) {
	cfg := Render("phone", "", "", []entity.SingBoxOut{
		{Tag: "香港01", Type: string(entity.OutboundProtocolShadowsocks), Server: "hk.example.com", ServerPort: 8388, Method: "aes-128-gcm", Password: "pass"},
	}, nil, nil, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector), Include: "香港"},
	}, []*entity.RuleSet{
		{Tag: "geosite-cn", RuleSetType: string(entity.RuleSetTypeLocal), Outbound: "general", Content: `{"version":1,"rules":[{"domain_suffix":["cn"]}]}`},
	})

	if !strings.Contains(cfg, "DOMAIN-SUFFIX,cn,general") {
		t.Errorf("expected expanded rule when system host empty:\n%s", cfg)
	}
}

// TestRenderSubscriptionPolicyPath 验证订阅源输出为携带 policy-path 的 select 策略组，
// 订阅节点不展开为 [Proxy] 行；节点分组通过 include-other-group 引用订阅策略组，
// 并把 Include/Exclude 关键字翻译为 policy-regex-filter。
func TestRenderSubscriptionPolicyPath(t *testing.T) {
	cfg := Render("phone", "", "", []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "manual-node",
			Server:     "manual.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "pass",
		},
	}, nil, []*entity.Subscribe{
		{Name: "provider-a", URL: "https://sub.example.com/a", Status: true, OutboundCacheDuration: 30},
		{Name: "provider-disabled", URL: "https://sub.example.com/b", Status: false},
		{Name: "provider-invisible", URL: "https://sub.example.com/c", Status: true, VisibleDevices: "tv"},
	}, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector), Include: "香港,manual", Exclude: "过期"},
	}, nil)

	// 订阅策略组：policy-path 指向订阅地址，缓存时长（分钟）映射为 update-interval（秒）。
	mustContain(t, cfg, "provider-a = select, policy-path=https://sub.example.com/a, update-interval=1800")
	// 禁用或对设备不可见的订阅不输出。
	mustNotContain(t, cfg, "provider-disabled")
	mustNotContain(t, cfg, "provider-invisible")
	// 手工节点仍展开为 [Proxy] 行。
	mustContain(t, cfg, "manual-node = ss, manual.example.com, 8388")
	// 节点分组引用订阅策略组并携带正则过滤。
	mustContain(t, cfg, "include-other-group=provider-a")
	mustContain(t, cfg, "policy-regex-filter=^(?!.*(过期)).*(香港|manual)")
}

// TestRenderSubscriptionOnlyGroup 验证没有任何手工节点命中时，
// 只要存在订阅策略组，节点分组仍然输出（成员完全来自 include-other-group）。
func TestRenderSubscriptionOnlyGroup(t *testing.T) {
	cfg := Render("phone", "", "", nil, nil, []*entity.Subscribe{
		{Name: "provider-a", URL: "https://sub.example.com/a", Status: true},
	}, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector), Include: "香港"},
	}, nil)

	mustContain(t, cfg, "provider-a = select, policy-path=https://sub.example.com/a")
	mustContain(t, cfg, "general = select, include-other-group=provider-a, policy-regex-filter=(香港)")
}
