package shadowrocket

import (
	"singboxconfig/entity"
	"strings"
	"testing"
)

func TestRenderProxyProtocols(t *testing.T) {
	cfg := Render("phone", entity.SingDNS{}, []entity.SingBoxOut{
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
			Type:          string(entity.OutboundProtocolShadowsocksR),
			Tag:           "ssr-node",
			Server:        "ssr.example.com",
			ServerPort:    2048,
			Method:        "chacha20-ietf",
			Password:      "ssr-pass",
			Protocol:      "auth_aes128_sha1",
			ProtocolParam: "42:abc",
			Obfs:          &entity.SingObfs{Value: "tls1.2_ticket_auth"},
			ObfsParam:     "download.example.com",
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
			Type:       string(entity.OutboundProtocolVLESS),
			Tag:        "vless-node",
			Server:     "vless.example.com",
			ServerPort: 443,
			UUID:       "00000000-0000-0000-0000-000000000001",
			Flow:       "xtls-rprx-vision",
			TLS: &entity.SingTLS{
				Enabled:    true,
				ServerName: "vless-sni.example.com",
				Utls:       &entity.SingUtls{Enabled: true, Fingerprint: "chrome"},
				Reality:    &entity.SingReality{Enabled: true, PublicKey: "pub-key", ShortID: "abcd"},
			},
		},
		{
			Type:       string(entity.OutboundProtocolHysteria2),
			Tag:        "hy2-node",
			Server:     "hy2.example.com",
			ServerPort: 8443,
			Password:   "hy2-pass",
			UpMbps:     50,
			DownMbps:   100,
			Obfs:       &entity.SingObfs{Type: "salamander", Value: "obfs-pass"},
			TLS:        &entity.SingTLS{Enabled: true, ServerName: "hy2-sni.example.com"},
		},
		{
			Type:                 string(entity.OutboundProtocolTUIC),
			Tag:                  "tuic-node",
			Server:               "tuic.example.com",
			ServerPort:           9443,
			UUID:                 "00000000-0000-0000-0000-000000000002",
			Password:             "tuic-pass",
			CongestionController: "bbr",
			UdpRelayMode:         "native",
			ZeroRttHandshake:     true,
			TLS:                  &entity.SingTLS{Enabled: true, ServerName: "tuic-sni.example.com"},
		},
	}, nil, nil)

	mustContain(t, cfg, "ss-node = ss, ss.example.com, 8388, encrypt-method=chacha20-ietf-poly1305, password=ss-pass, udp-relay=true, obfs=tls, obfs-host=cdn.example.com, obfs-uri=/")
	mustContain(t, cfg, "ssr-node = ssr, ssr.example.com, 2048, encrypt-method=chacha20-ietf, password=ssr-pass, protocol=auth_aes128_sha1, udp-relay=true, protocol-param=42:abc, obfs=tls1.2_ticket_auth, obfs-param=download.example.com")
	mustContain(t, cfg, "trojan-node = trojan, trojan.example.com, 443, password=trojan-pass, udp-relay=true, tls=true, sni=sni.example.com, skip-cert-verify=true")
	mustContain(t, cfg, "vmess-node = vmess, vmess.example.com, 443, username=0233d11c-15a4-47d3-ade3-48ffca0ce119, udp-relay=true, encrypt-method=aes-128-gcm, tls=true, sni=vmess-sni.example.com, ws=true, ws-path=/ws, ws-headers=Host:host.example.com")
	mustContain(t, cfg, "vless-node = vless, vless.example.com, 443, username=00000000-0000-0000-0000-000000000001, udp-relay=true, flow=xtls-rprx-vision, tls=true, sni=vless-sni.example.com, fingerprint=chrome, reality=true, public-key=pub-key, short-id=abcd")
	mustContain(t, cfg, "hy2-node = hysteria2, hy2.example.com, 8443, password=hy2-pass, tls=true, sni=hy2-sni.example.com, up=50 Mbps, down=100 Mbps, obfs=salamander, obfs-password=obfs-pass")
	mustContain(t, cfg, "tuic-node = tuic, tuic.example.com, 9443, uuid=00000000-0000-0000-0000-000000000002, password=tuic-pass, tls=true, sni=tuic-sni.example.com, congestion-controller=bbr, udp-relay-mode=native, zero-rtt-handshake=true")
}

func TestRenderSkipsUnsupportedProtocolAndKeepsGroupReferencesValid(t *testing.T) {
	cfg := Render("phone", entity.SingDNS{}, []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "香港01",
			Server:     "hk.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "pass",
		},
		{
			Type:       string(entity.OutboundProtocolHysteria),
			Tag:        "香港Hysteria",
			Server:     "hy.example.com",
			ServerPort: 443,
			AuthStr:    "auth",
		},
	}, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector), Include: "香港"},
	}, nil)

	mustContain(t, cfg, "香港01 = ss, hk.example.com, 8388")
	mustNotContain(t, cfg, "香港Hysteria =")
	mustContain(t, cfg, "general = select, 香港01")
	mustNotContain(t, cfg, "general = select, 香港01, 香港Hysteria")
}

func TestRenderExpandsRuleSets(t *testing.T) {
	ruleContent := `{"rules":[{"domain":["exact.example.com"],"domain_suffix":["example.com"],"domain_keyword":["video"],"domain_regex":["^api\\."],"ip_cidr":["192.168.0.0/16","2001:db8::/32"],"geoip":["cn"]}]}`
	cfg := Render("phone", entity.SingDNS{}, []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "proxy-node",
			Server:     "proxy.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "pass",
		},
	}, []*entity.NodeGroup{
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
	mustContain(t, cfg, "GEOIP,CN,general")
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
