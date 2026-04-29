package protocol

import (
	"testing"
)

// 以下为脱敏后的虚构 VLESS 节点示例（UUID、服务器地址、公钥、ShortID 均为占位数据）：
//
//	vless://00000000-0000-0000-0000-000000000001@node-jp.example.com:20105?mode=multi&security=reality&encryption=none&type=tcp&flow=xtls-rprx-vision&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=aabb00000001&sni=sni.example.com&servername=sni.example.com&spx=%2F&fp=chrome#日本02
//	vless://00000000-0000-0000-0000-000000000001@node-br.example.com:5518?mode=multi&security=reality&encryption=none&type=tcp&flow=xtls-rprx-vision&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=aabb0002&sni=sni.example.com&servername=sni.example.com&spx=%2F&fp=chrome#巴西01-直连
//	vless://00000000-0000-0000-0000-000000000001@node-sg.example.com:5518?mode=multi&security=reality&encryption=none&type=tcp&flow=xtls-rprx-vision&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=aabb000000000003&sni=sni.example.com&servername=sni.example.com&spx=%2F&fp=chrome#新加坡01-直连

// ─── DecodeVlessUrl 测试 ─────────────────────────────────────────────────────

// Test_DecodeVlessUrl_Japan 测试日本节点（Reality + TCP + xtls-rprx-vision）的 URL 解析。
func Test_DecodeVlessUrl_Japan(t *testing.T) {
	urlStr := "vless://00000000-0000-0000-0000-000000000001@node-jp.example.com:20105?mode=multi&security=reality&encryption=none&type=tcp&flow=xtls-rprx-vision&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=aabb00000001&sni=sni.example.com&servername=sni.example.com&spx=%2F&fp=chrome#%E6%97%A5%E6%9C%AC02"

	node, err := DecodeVlessUrl(urlStr)
	if err != nil {
		t.Fatalf("DecodeVlessUrl failed: %v", err)
	}

	// 验证各字段
	if node.UUID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("UUID = %v, want 00000000-0000-0000-0000-000000000001", node.UUID)
	}
	if node.Host != "node-jp.example.com" {
		t.Errorf("Host = %v, want node-jp.example.com", node.Host)
	}
	if node.Port != 20105 {
		t.Errorf("Port = %v, want 20105", node.Port)
	}
	if node.Security != "reality" {
		t.Errorf("Security = %v, want reality", node.Security)
	}
	if node.Encryption != "none" {
		t.Errorf("Encryption = %v, want none", node.Encryption)
	}
	if node.Type != "tcp" {
		t.Errorf("Type = %v, want tcp", node.Type)
	}
	if node.Flow != "xtls-rprx-vision" {
		t.Errorf("Flow = %v, want xtls-rprx-vision", node.Flow)
	}
	if node.PublicKey != "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" {
		t.Errorf("PublicKey = %v, want AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", node.PublicKey)
	}
	if node.ShortID != "aabb00000001" {
		t.Errorf("ShortID = %v, want aabb00000001", node.ShortID)
	}
	// sni 参数优先于 servername
	if node.SNI != "sni.example.com" {
		t.Errorf("SNI = %v, want sni.example.com", node.SNI)
	}
	if node.Fingerprint != "chrome" {
		t.Errorf("Fingerprint = %v, want chrome", node.Fingerprint)
	}
	// cleanTag 会去除标点，但中文和数字保留
	if node.Tag != "日本02" {
		t.Errorf("Tag = %v, want 日本02", node.Tag)
	}
}

// Test_DecodeVlessUrl_Brazil 测试巴西节点（Reality + TCP），验证标签中 "-" 被 cleanTag 去除。
func Test_DecodeVlessUrl_Brazil(t *testing.T) {
	urlStr := "vless://00000000-0000-0000-0000-000000000001@node-br.example.com:5518?mode=multi&security=reality&encryption=none&type=tcp&flow=xtls-rprx-vision&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=aabb0002&sni=sni.example.com&servername=sni.example.com&spx=%2F&fp=chrome#%E5%B7%B4%E8%A5%BF01-%E7%9B%B4%E8%BF%9E"

	node, err := DecodeVlessUrl(urlStr)
	if err != nil {
		t.Fatalf("DecodeVlessUrl failed: %v", err)
	}

	if node.Host != "node-br.example.com" {
		t.Errorf("Host = %v, want node-br.example.com", node.Host)
	}
	if node.Port != 5518 {
		t.Errorf("Port = %v, want 5518", node.Port)
	}
	if node.ShortID != "aabb0002" {
		t.Errorf("ShortID = %v, want aabb0002", node.ShortID)
	}
	// "巴西01-直连" 经 cleanTag 后，"-" 被移除，结果为 "巴西01直连"
	if node.Tag != "巴西01直连" {
		t.Errorf("Tag = %v, want 巴西01直连", node.Tag)
	}
}

// Test_DecodeVlessUrl_Singapore 测试新加坡节点（Reality + TCP），验证较长的 ShortID。
func Test_DecodeVlessUrl_Singapore(t *testing.T) {
	urlStr := "vless://00000000-0000-0000-0000-000000000001@node-sg.example.com:5518?mode=multi&security=reality&encryption=none&type=tcp&flow=xtls-rprx-vision&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=aabb000000000003&sni=sni.example.com&servername=sni.example.com&spx=%2F&fp=chrome#%E6%96%B0%E5%8A%A0%E5%9D%A101-%E7%9B%B4%E8%BF%9E"

	node, err := DecodeVlessUrl(urlStr)
	if err != nil {
		t.Fatalf("DecodeVlessUrl failed: %v", err)
	}

	if node.Host != "node-sg.example.com" {
		t.Errorf("Host = %v, want node-sg.example.com", node.Host)
	}
	if node.Port != 5518 {
		t.Errorf("Port = %v, want 5518", node.Port)
	}
	// 较长的 ShortID
	if node.ShortID != "aabb000000000003" {
		t.Errorf("ShortID = %v, want aabb000000000003", node.ShortID)
	}
	// "新加坡01-直连" 经 cleanTag 后，"-" 被移除
	if node.Tag != "新加坡01直连" {
		t.Errorf("Tag = %v, want 新加坡01直连", node.Tag)
	}
}

// Test_DecodeVlessUrl_ServernameOnly 测试仅有 servername 参数（无 sni）时，应回退读取 servername。
func Test_DecodeVlessUrl_ServernameOnly(t *testing.T) {
	urlStr := "vless://aaaabbbb-cccc-dddd-eeee-ffffffffffff@example.com:443?security=tls&servername=sni.example.com&fp=firefox#TLS节点"

	node, err := DecodeVlessUrl(urlStr)
	if err != nil {
		t.Fatalf("DecodeVlessUrl failed: %v", err)
	}

	// sni 参数为空，应读取 servername
	if node.SNI != "sni.example.com" {
		t.Errorf("SNI = %v, want sni.example.com", node.SNI)
	}
	if node.Security != "tls" {
		t.Errorf("Security = %v, want tls", node.Security)
	}
}

// Test_DecodeVlessUrl_NoFragment 测试没有标签（fragment）时回退为默认标签。
func Test_DecodeVlessUrl_NoFragment(t *testing.T) {
	urlStr := "vless://aaaabbbb-cccc-dddd-eeee-ffffffffffff@example.com:443?security=none"

	node, err := DecodeVlessUrl(urlStr)
	if err != nil {
		t.Fatalf("DecodeVlessUrl failed: %v", err)
	}

	if node.Tag != "vless-node" {
		t.Errorf("Tag = %v, want vless-node", node.Tag)
	}
}

// Test_DecodeVlessUrl_InvalidScheme 测试非 vless:// 协议头应返回错误。
func Test_DecodeVlessUrl_InvalidScheme(t *testing.T) {
	_, err := DecodeVlessUrl("vmess://example.com")
	if err == nil {
		t.Error("expected error for invalid scheme, got nil")
	}
}

// Test_DecodeVlessUrl_MissingUUID 测试缺少 UUID 时应返回错误。
func Test_DecodeVlessUrl_MissingUUID(t *testing.T) {
	_, err := DecodeVlessUrl("vless://example.com:443")
	if err == nil {
		t.Error("expected error for missing UUID, got nil")
	}
}

// Test_DecodeVlessUrl_InvalidPort 测试非法端口应返回错误。
func Test_DecodeVlessUrl_InvalidPort(t *testing.T) {
	_, err := DecodeVlessUrl("vless://uuid@example.com:notaport")
	if err == nil {
		t.Error("expected error for invalid port, got nil")
	}
}

// ─── ConvertVlessToSingBox 测试 ──────────────────────────────────────────────

// Test_ConvertVlessToSingBox_Reality 测试 Reality 安全层（含 uTLS 指纹和 xtls-rprx-vision 流控）的转换结果。
func Test_ConvertVlessToSingBox_Reality(t *testing.T) {
	node := &VlessNode{
		UUID:        "00000000-0000-0000-0000-000000000001",
		Host:        "node-jp.example.com",
		Port:        20105,
		Security:    "reality",
		Encryption:  "none",
		Type:        "tcp",
		Flow:        "xtls-rprx-vision",
		PublicKey:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		ShortID:     "aabb00000001",
		SNI:         "sni.example.com",
		Fingerprint: "chrome",
		Tag:         "日本02",
	}

	out, err := ConvertVlessToSingBox(node)
	if err != nil {
		t.Fatalf("ConvertVlessToSingBox failed: %v", err)
	}

	// 验证基础字段
	if out.Type != "vless" {
		t.Errorf("Type = %v, want vless", out.Type)
	}
	if out.Tag != "日本02" {
		t.Errorf("Tag = %v, want 日本02", out.Tag)
	}
	if out.Server != "node-jp.example.com" {
		t.Errorf("Server = %v, want node-jp.example.com", out.Server)
	}
	if out.ServerPort != 20105 {
		t.Errorf("ServerPort = %v, want 20105", out.ServerPort)
	}
	if out.UUID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("UUID = %v, want 00000000-0000-0000-0000-000000000001", out.UUID)
	}
	if out.Flow != "xtls-rprx-vision" {
		t.Errorf("Flow = %v, want xtls-rprx-vision", out.Flow)
	}

	// 验证 TLS 配置
	if out.TLS == nil {
		t.Fatal("TLS is nil, expected Reality TLS config")
	}
	if !out.TLS.Enabled {
		t.Error("TLS.Enabled = false, want true")
	}
	if out.TLS.ServerName != "sni.example.com" {
		t.Errorf("TLS.ServerName = %v, want sni.example.com", out.TLS.ServerName)
	}

	// 验证 Reality 配置
	if out.TLS.Reality == nil {
		t.Fatal("TLS.Reality is nil")
	}
	if !out.TLS.Reality.Enabled {
		t.Error("TLS.Reality.Enabled = false, want true")
	}
	if out.TLS.Reality.PublicKey != "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" {
		t.Errorf("TLS.Reality.PublicKey = %v, want AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", out.TLS.Reality.PublicKey)
	}
	if out.TLS.Reality.ShortID != "aabb00000001" {
		t.Errorf("TLS.Reality.ShortID = %v, want aabb00000001", out.TLS.Reality.ShortID)
	}

	// 验证 uTLS 指纹配置
	if out.TLS.Utls == nil {
		t.Fatal("TLS.Utls is nil")
	}
	if !out.TLS.Utls.Enabled {
		t.Error("TLS.Utls.Enabled = false, want true")
	}
	if out.TLS.Utls.Fingerprint != "chrome" {
		t.Errorf("TLS.Utls.Fingerprint = %v, want chrome", out.TLS.Utls.Fingerprint)
	}

	// TCP 传输不产生 transport 字段
	if out.Transport != nil && out.Transport.Type != "" {
		t.Errorf("Transport.Type = %v, want empty (TCP needs no transport)", out.Transport.Type)
	}
}

// Test_ConvertVlessToSingBox_TLS 测试标准 TLS 安全层（无 Reality）的转换结果。
func Test_ConvertVlessToSingBox_TLS(t *testing.T) {
	node := &VlessNode{
		UUID:        "aaaabbbb-cccc-dddd-eeee-ffffffffffff",
		Host:        "example.com",
		Port:        443,
		Security:    "tls",
		Type:        "tcp",
		SNI:         "sni.example.com",
		Fingerprint: "firefox",
		Tag:         "TLS节点",
	}

	out, err := ConvertVlessToSingBox(node)
	if err != nil {
		t.Fatalf("ConvertVlessToSingBox failed: %v", err)
	}

	if out.Type != "vless" {
		t.Errorf("Type = %v, want vless", out.Type)
	}
	if out.TLS == nil {
		t.Fatal("TLS is nil for tls security")
	}
	if !out.TLS.Enabled {
		t.Error("TLS.Enabled = false, want true")
	}
	if out.TLS.ServerName != "sni.example.com" {
		t.Errorf("TLS.ServerName = %v, want sni.example.com", out.TLS.ServerName)
	}
	// TLS 模式不应该有 Reality 配置
	if out.TLS.Reality != nil {
		t.Error("TLS.Reality should be nil for plain TLS security")
	}
	if out.TLS.Utls == nil || out.TLS.Utls.Fingerprint != "firefox" {
		t.Errorf("TLS.Utls.Fingerprint = %v, want firefox", func() string {
			if out.TLS.Utls == nil {
				return "<nil>"
			}
			return out.TLS.Utls.Fingerprint
		}())
	}
}

// Test_ConvertVlessToSingBox_NoTLS 测试无安全层（security=none 或空）时不生成 TLS 配置。
func Test_ConvertVlessToSingBox_NoTLS(t *testing.T) {
	node := &VlessNode{
		UUID:     "aaaabbbb-cccc-dddd-eeee-ffffffffffff",
		Host:     "example.com",
		Port:     80,
		Security: "none",
		Type:     "tcp",
		Tag:      "NoTLS节点",
	}

	out, err := ConvertVlessToSingBox(node)
	if err != nil {
		t.Fatalf("ConvertVlessToSingBox failed: %v", err)
	}

	if out.TLS != nil {
		t.Error("TLS should be nil for security=none")
	}
}

// Test_ConvertVlessToSingBox_WebSocket 测试 WebSocket 传输层的转换结果。
func Test_ConvertVlessToSingBox_WebSocket(t *testing.T) {
	node := &VlessNode{
		UUID:       "aaaabbbb-cccc-dddd-eeee-ffffffffffff",
		Host:       "ws.example.com",
		Port:       443,
		Security:   "tls",
		Type:       "ws",
		SNI:        "ws.example.com",
		Path:       "/ws",
		HostHeader: "ws.example.com",
		Tag:        "WS节点",
	}

	out, err := ConvertVlessToSingBox(node)
	if err != nil {
		t.Fatalf("ConvertVlessToSingBox failed: %v", err)
	}

	// 验证 WebSocket transport
	if out.Transport == nil {
		t.Fatal("Transport is nil for ws type")
	}
	if out.Transport.Type != "ws" {
		t.Errorf("Transport.Type = %v, want ws", out.Transport.Type)
	}
	if out.Transport.Path != "/ws" {
		t.Errorf("Transport.Path = %v, want /ws", out.Transport.Path)
	}
	if len(out.Transport.Headers["Host"]) == 0 || out.Transport.Headers["Host"][0] != "ws.example.com" {
		t.Errorf("Transport.Headers[Host] = %v, want [ws.example.com]", out.Transport.Headers["Host"])
	}
}

// Test_ConvertVlessToSingBox_WebSocket_DefaultPath 测试 WebSocket 路径为空时应回退为 "/"。
func Test_ConvertVlessToSingBox_WebSocket_DefaultPath(t *testing.T) {
	node := &VlessNode{
		UUID:     "aaaabbbb-cccc-dddd-eeee-ffffffffffff",
		Host:     "ws.example.com",
		Port:     443,
		Security: "tls",
		Type:     "ws",
		Tag:      "WS空路径节点",
	}

	out, err := ConvertVlessToSingBox(node)
	if err != nil {
		t.Fatalf("ConvertVlessToSingBox failed: %v", err)
	}

	if out.Transport == nil {
		t.Fatal("Transport is nil")
	}
	if out.Transport.Path != "/" {
		t.Errorf("Transport.Path = %v, want / (default)", out.Transport.Path)
	}
}

// Test_ConvertVlessToSingBox_GRPC 测试 gRPC 传输层的转换结果。
func Test_ConvertVlessToSingBox_GRPC(t *testing.T) {
	node := &VlessNode{
		UUID:        "aaaabbbb-cccc-dddd-eeee-ffffffffffff",
		Host:        "grpc.example.com",
		Port:        443,
		Security:    "tls",
		Type:        "grpc",
		SNI:         "grpc.example.com",
		ServiceName: "MyService",
		Tag:         "gRPC节点",
	}

	out, err := ConvertVlessToSingBox(node)
	if err != nil {
		t.Fatalf("ConvertVlessToSingBox failed: %v", err)
	}

	if out.Transport == nil {
		t.Fatal("Transport is nil for grpc type")
	}
	if out.Transport.Type != "grpc" {
		t.Errorf("Transport.Type = %v, want grpc", out.Transport.Type)
	}
	if out.Transport.ServiceName != "MyService" {
		t.Errorf("Transport.ServiceName = %v, want MyService", out.Transport.ServiceName)
	}
}

// Test_ConvertVlessToSingBox_Nil 测试传入 nil 时应返回错误。
func Test_ConvertVlessToSingBox_Nil(t *testing.T) {
	_, err := ConvertVlessToSingBox(nil)
	if err == nil {
		t.Error("expected error for nil input, got nil")
	}
}

// ─── DecodeVlessUrlToSingBox 端到端测试 ──────────────────────────────────────

// Test_DecodeVlessUrlToSingBox_Japan 端到端测试：直接从日本节点 URL 生成 sing-box 出站配置。
func Test_DecodeVlessUrlToSingBox_Japan(t *testing.T) {
	urlStr := "vless://00000000-0000-0000-0000-000000000001@node-jp.example.com:20105?mode=multi&security=reality&encryption=none&type=tcp&flow=xtls-rprx-vision&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=aabb00000001&sni=sni.example.com&servername=sni.example.com&spx=%2F&fp=chrome#%E6%97%A5%E6%9C%AC02"

	out, err := DecodeVlessUrlToSingBox(urlStr)
	if err != nil {
		t.Fatalf("DecodeVlessUrlToSingBox failed: %v", err)
	}

	// 验证关键字段
	if out.Type != "vless" {
		t.Errorf("Type = %v, want vless", out.Type)
	}
	if out.Tag != "日本02" {
		t.Errorf("Tag = %v, want 日本02", out.Tag)
	}
	if out.Server != "node-jp.example.com" {
		t.Errorf("Server = %v, want node-jp.example.com", out.Server)
	}
	if out.ServerPort != 20105 {
		t.Errorf("ServerPort = %v, want 20105", out.ServerPort)
	}
	if out.UUID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("UUID = %v, want 00000000-0000-0000-0000-000000000001", out.UUID)
	}
	if out.Flow != "xtls-rprx-vision" {
		t.Errorf("Flow = %v, want xtls-rprx-vision", out.Flow)
	}
	if out.TLS == nil || !out.TLS.Enabled {
		t.Error("TLS should be enabled")
	}
	if out.TLS.Reality == nil || !out.TLS.Reality.Enabled {
		t.Error("Reality should be enabled")
	}
	if out.TLS.Reality.PublicKey != "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" {
		t.Errorf("PublicKey = %v", out.TLS.Reality.PublicKey)
	}
	if out.TLS.Reality.ShortID != "aabb00000001" {
		t.Errorf("ShortID = %v, want aabb00000001", out.TLS.Reality.ShortID)
	}
	if out.TLS.Utls == nil || out.TLS.Utls.Fingerprint != "chrome" {
		t.Errorf("Fingerprint = %v, want chrome", func() string {
			if out.TLS.Utls == nil {
				return "<nil>"
			}
			return out.TLS.Utls.Fingerprint
		}())
	}
}

// Test_DecodeVlessUrlToSingBox_Brazil 端到端测试：巴西节点 URL。
func Test_DecodeVlessUrlToSingBox_Brazil(t *testing.T) {
	urlStr := "vless://00000000-0000-0000-0000-000000000001@node-br.example.com:5518?mode=multi&security=reality&encryption=none&type=tcp&flow=xtls-rprx-vision&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=aabb0002&sni=sni.example.com&servername=sni.example.com&spx=%2F&fp=chrome#%E5%B7%B4%E8%A5%BF01-%E7%9B%B4%E8%BF%9E"

	out, err := DecodeVlessUrlToSingBox(urlStr)
	if err != nil {
		t.Fatalf("DecodeVlessUrlToSingBox failed: %v", err)
	}

	if out.Server != "node-br.example.com" {
		t.Errorf("Server = %v, want node-br.example.com", out.Server)
	}
	if out.ServerPort != 5518 {
		t.Errorf("ServerPort = %v, want 5518", out.ServerPort)
	}
	if out.TLS.Reality.ShortID != "aabb0002" {
		t.Errorf("ShortID = %v, want aabb0002", out.TLS.Reality.ShortID)
	}
	// "-" 被 cleanTag 去除
	if out.Tag != "巴西01直连" {
		t.Errorf("Tag = %v, want 巴西01直连", out.Tag)
	}
}

// Test_DecodeVlessUrlToSingBox_Singapore 端到端测试：新加坡节点 URL（较长的 ShortID）。
func Test_DecodeVlessUrlToSingBox_Singapore(t *testing.T) {
	urlStr := "vless://00000000-0000-0000-0000-000000000001@node-sg.example.com:5518?mode=multi&security=reality&encryption=none&type=tcp&flow=xtls-rprx-vision&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=aabb000000000003&sni=sni.example.com&servername=sni.example.com&spx=%2F&fp=chrome#%E6%96%B0%E5%8A%A0%E5%9D%A101-%E7%9B%B4%E8%BF%9E"

	out, err := DecodeVlessUrlToSingBox(urlStr)
	if err != nil {
		t.Fatalf("DecodeVlessUrlToSingBox failed: %v", err)
	}

	if out.Server != "node-sg.example.com" {
		t.Errorf("Server = %v, want node-sg.example.com", out.Server)
	}
	if out.TLS.Reality.ShortID != "aabb000000000003" {
		t.Errorf("ShortID = %v, want aabb000000000003", out.TLS.Reality.ShortID)
	}
	if out.Tag != "新加坡01直连" {
		t.Errorf("Tag = %v, want 新加坡01直连", out.Tag)
	}
}

// Test_DecodeVlessUrlToSingBox_Error 端到端测试：非法 URL 应返回错误。
func Test_DecodeVlessUrlToSingBox_Error(t *testing.T) {
	_, err := DecodeVlessUrlToSingBox("not-a-valid-url")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}
