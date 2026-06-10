package surge

import (
	"singboxconfig/entity"
	"strings"
	"testing"
)

func TestRenderProxyProtocols(t *testing.T) {
	cfg := Render("phone", []entity.SingBoxOut{
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
	}, nil, nil, nil)

	mustContain(t, cfg, "ss-node = ss, ss.example.com, 8388, encrypt-method=chacha20-ietf-poly1305, password=ss-pass, udp-relay=true, obfs=tls, obfs-host=cdn.example.com, obfs-uri=/")
	mustContain(t, cfg, "trojan-node = trojan, trojan.example.com, 443, password=trojan-pass, udp-relay=true, tls=true, sni=sni.example.com, skip-cert-verify=true")
	mustContain(t, cfg, "vmess-node = vmess, vmess.example.com, 443, username=0233d11c-15a4-47d3-ade3-48ffca0ce119, encrypt-method=aes-128-gcm, tls=true, sni=vmess-sni.example.com, ws=true, ws-path=/ws, ws-headers=Host:host.example.com")
	mustContain(t, cfg, "http-node = http, http.example.com, 8080, username=http-user, password=http-pass")
	mustContain(t, cfg, "https-node = https, https.example.com, 443, username=https-user, password=https-pass, tls=true, sni=https-sni.example.com")
}

func TestRenderWireGuardEndpoints(t *testing.T) {
	cfg := Render("phone", []entity.SingBoxOut{
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
	}, []*entity.NodeGroup{
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
	cfg := Render("phone", []entity.SingBoxOut{
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
	}, nil, []*entity.NodeGroup{
		{Tag: "general", GroupType: string(entity.NodeGroupTypeSelector), Include: "香港"},
	}, nil)

	mustContain(t, cfg, "香港01 = ss, hk.example.com, 8388")
	mustNotContain(t, cfg, "香港VLESS =")
	mustContain(t, cfg, "general = select, 香港01")
	mustNotContain(t, cfg, "general = select, 香港01, 香港VLESS")
}

func TestRenderExpandsRuleSets(t *testing.T) {
	ruleContent := `{"rules":[{"domain":["exact.example.com"],"domain_suffix":["example.com"],"domain_keyword":["video"],"domain_regex":["^api\\."],"ip_cidr":["192.168.0.0/16","2001:db8::/32"],"geoip":["cn"]}]}`
	cfg := Render("phone", []entity.SingBoxOut{
		{
			Type:       string(entity.OutboundProtocolShadowsocks),
			Tag:        "proxy-node",
			Server:     "proxy.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "pass",
		},
	}, nil, []*entity.NodeGroup{
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
