package protocol

import (
	"encoding/base64"
	"testing"
)

func TestDecodeVmessUrl(t *testing.T) {
	t.Run("基础URL解析", func(t *testing.T) {
		// 使用代码注释中的示例 URL
		inputURL := "vmess://eyJ2IjoiMiIsInBzIjoiSEsuXHU5OTk5XHU2ZTJmLkMgfCBcdTlhZDhcdTkwMWYuMnhcdTUwMGRcdTczODciLCJhZGQiOiJ4Y3RyYW5zZmVyMXN0LmNva2VjbG91ZC50b3AiLCJwb3J0IjoiMzg2MzAiLCJpZCI6IjNmYzM2NWYyLTk0MGMtNGUwNy05Y2Y3LWExMGM1N2JlODU3YSIsImFpZCI6IjAiLCJuZXQiOiJ0Y3AiLCJ0eXBlIjoibm9uZSIsImhvc3QiOiIiLCJwYXRoIjoiIiwidGxzIjoiIn0="

		expectedItem := &VmessNode{
			Version: "2",
			Remarks: "HK.香港.C | 高速.2x倍率",
			Address: "xctransfer1st.cokecloud.top",
			Port:    "38630",
			ID:      "3fc365f2-940c-4e07-9cf7-a10c57be857a",
			AlterID: "0",
			Network: "tcp",
			Type:    "none",
			Host:    "",
			Path:    "",
			TLS:     "",
		}

		item, err := DecodeVmessUrl(inputURL)
		if err != nil {
			t.Fatalf("DecodeVmessUrl failed: %v", err)
		}

		// 验证所有字段
		if item.Version != expectedItem.Version {
			t.Errorf("Version = %v, want %v", item.Version, expectedItem.Version)
		}
		if item.Remarks != expectedItem.Remarks {
			t.Errorf("Remarks = %v, want %v", item.Remarks, expectedItem.Remarks)
		}
		if item.Address != expectedItem.Address {
			t.Errorf("Address = %v, want %v", item.Address, expectedItem.Address)
		}
		if item.Port != expectedItem.Port {
			t.Errorf("Port = %v, want %v", item.Port, expectedItem.Port)
		}
		if item.ID != expectedItem.ID {
			t.Errorf("ID = %v, want %v", item.ID, expectedItem.ID)
		}
		if item.AlterID != expectedItem.AlterID {
			t.Errorf("AlterID = %v, want %v", item.AlterID, expectedItem.AlterID)
		}
		if item.Network != expectedItem.Network {
			t.Errorf("Network = %v, want %v", item.Network, expectedItem.Network)
		}
		if item.Type != expectedItem.Type {
			t.Errorf("Type = %v, want %v", item.Type, expectedItem.Type)
		}
		if item.Host != expectedItem.Host {
			t.Errorf("Host = %v, want %v", item.Host, expectedItem.Host)
		}
		if item.Path != expectedItem.Path {
			t.Errorf("Path = %v, want %v", item.Path, expectedItem.Path)
		}
		if item.TLS != expectedItem.TLS {
			t.Errorf("TLS = %v, want %v", item.TLS, expectedItem.TLS)
		}
	})

	t.Run("WebSocket协议解析", func(t *testing.T) {
		// 测试 WebSocket 传输协议的 vmess URL
		wsConfig := `{"v":"2","ps":"WS节点","add":"example.com","port":"443","id":"test-uuid","aid":"0","net":"ws","type":"none","host":"example.com","path":"/ws","tls":"tls"}`
		wsURL := "vmess://" + encodeBase64(wsConfig)

		item, err := DecodeVmessUrl(wsURL)
		if err != nil {
			t.Fatalf("DecodeVmessUrl failed: %v", err)
		}

		if item.Network != "ws" {
			t.Errorf("Network = %v, want %v", item.Network, "ws")
		}
		if item.Path != "/ws" {
			t.Errorf("Path = %v, want %v", item.Path, "/ws")
		}
		if item.TLS != "tls" {
			t.Errorf("TLS = %v, want %v", item.TLS, "tls")
		}
		if item.Host != "example.com" {
			t.Errorf("Host = %v, want %v", item.Host, "example.com")
		}
	})

	t.Run("无效协议处理", func(t *testing.T) {
		invalidURL := "http://example.com"
		_, err := DecodeVmessUrl(invalidURL)
		if err == nil {
			t.Error("Expected error for invalid scheme, but got nil")
		}
	})

	t.Run("无效Base64处理", func(t *testing.T) {
		invalidURL := "vmess://invalid-base64!"
		_, err := DecodeVmessUrl(invalidURL)
		if err == nil {
			t.Error("Expected error for invalid base64, but got nil")
		}
	})
}

func TestConvertVmessToSingBox(t *testing.T) {
	t.Run("基础转换", func(t *testing.T) {
		node := &VmessNode{
			Version: "2",
			Remarks: "Test Node",
			Address: "example.com",
			Port:    "443",
			ID:      "test-uuid-1234",
			AlterID: "0",
			Network: "tcp",
			Type:    "none",
			Host:    "",
			Path:    "",
			TLS:     "",
		}

		out, err := ConvertVmessToSingBox(node)
		if err != nil {
			t.Fatalf("ConvertVmessToSingBox failed: %v", err)
		}

		// 验证基本字段
		if out.Type != "vmess" {
			t.Errorf("Type = %v, want %v", out.Type, "vmess")
		}
		if out.Tag != "TestNode" { // cleanTag 会移除空格
			t.Errorf("Tag = %v, want %v", out.Tag, "TestNode")
		}
		if out.Server != node.Address {
			t.Errorf("Server = %v, want %v", out.Server, node.Address)
		}
		if out.ServerPort != 443 {
			t.Errorf("ServerPort = %v, want %v", out.ServerPort, 443)
		}
		if out.UUID != node.ID {
			t.Errorf("UUID = %v, want %v", out.UUID, node.ID)
		}
		if out.AlterID != 0 {
			t.Errorf("AlterID = %v, want %v", out.AlterID, 0)
		}
		if out.Security != "auto" {
			t.Errorf("Security = %v, want %v", out.Security, "auto")
		}
	})

	t.Run("TLS配置转换", func(t *testing.T) {
		node := &VmessNode{
			Version: "2",
			Remarks: "TLS Node",
			Address: "example.com",
			Port:    "443",
			ID:      "test-uuid-1234",
			AlterID: "0",
			Network: "ws",
			Type:    "none",
			Host:    "example.com",
			Path:    "/path",
			TLS:     "tls",
		}

		out, err := ConvertVmessToSingBox(node)
		if err != nil {
			t.Fatalf("ConvertVmessToSingBox failed: %v", err)
		}

		// 验证 TLS 配置
		if out.TLS == nil {
			t.Fatal("TLS config is nil")
		}
		if !out.TLS.Enabled {
			t.Errorf("TLS.Enabled = %v, want %v", out.TLS.Enabled, true)
		}
		if out.TLS.ServerName != "example.com" {
			t.Errorf("TLS.ServerName = %v, want %v", out.TLS.ServerName, "example.com")
		}

		// 验证 WebSocket 传输配置
		if out.Transport == nil {
			t.Fatal("Transport config is nil")
		}
		if out.Transport.Type != "ws" {
			t.Errorf("Transport.Type = %v, want %v", out.Transport.Type, "ws")
		}
		if out.Transport.Path != "/path" {
			t.Errorf("Transport.Path = %v, want %v", out.Transport.Path, "/path")
		}
		if out.Transport.Headers["Host"][0] != "example.com" {
			t.Errorf("Transport.Headers[Host] = %v, want %v", out.Transport.Headers["Host"][0], "example.com")
		}
	})

	t.Run("默认端口处理", func(t *testing.T) {
		tests := []struct {
			name     string
			port     string
			tls      string
			expected int
		}{
			{"TLS默认端口", "", "tls", 443},
			{"非TLS默认端口", "0", "", 80},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				node := &VmessNode{
					Version: "2",
					Remarks: "Port Test",
					Address: "example.com",
					Port:    tt.port,
					ID:      "test-uuid-1234",
					AlterID: "0",
					Network: "tcp",
					Type:    "none",
					TLS:     tt.tls,
				}

				out, err := ConvertVmessToSingBox(node)
				if err != nil {
					t.Fatalf("ConvertVmessToSingBox failed: %v", err)
				}

				if out.ServerPort != tt.expected {
					t.Errorf("ServerPort = %v, want %v", out.ServerPort, tt.expected)
				}
			})
		}
	})

	t.Run("传输协议支持", func(t *testing.T) {
		tests := []struct {
			name        string
			network     string
			path        string
			host        string
			expectType  string
			expectPath  string
			expectHosts []string
		}{
			{
				name:       "gRPC协议",
				network:    "grpc",
				path:       "MyService",
				expectType: "grpc",
			},
			{
				name:       "gRPC默认服务",
				network:    "grpc",
				path:       "",
				expectType: "grpc",
			},
			{
				name:        "HTTP2协议",
				network:     "h2",
				path:        "/h2path",
				host:        "host1.com,host2.com",
				expectType:  "http",
				expectPath:  "/h2path",
				expectHosts: []string{"host1.com", "host2.com"},
			},
			{
				name:       "TCP HTTP伪装",
				network:    "tcp",
				path:       "/fake",
				host:       "fake.host.com",
				expectType: "http",
				expectPath: "/fake",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				node := &VmessNode{
					Version: "2",
					Remarks: tt.name,
					Address: "example.com",
					Port:    "443",
					ID:      "test-uuid-1234",
					AlterID: "0",
					Network: tt.network,
					Type: func() string {
						if tt.network == "tcp" {
							return "http"
						}
						return "none"
					}(),
					Host: tt.host,
					Path: tt.path,
					TLS:  "tls",
				}

				out, err := ConvertVmessToSingBox(node)
				if err != nil {
					t.Fatalf("ConvertVmessToSingBox failed: %v", err)
				}

				if tt.network == "grpc" {
					if out.Transport == nil {
						t.Fatal("Transport config is nil")
					}
					if out.Transport.Type != tt.expectType {
						t.Errorf("Transport.Type = %v, want %v", out.Transport.Type, tt.expectType)
					}
					expectedService := tt.path
					if expectedService == "" {
						expectedService = "GunService"
					}
					if out.Transport.ServiceName != expectedService {
						t.Errorf("Transport.ServiceName = %v, want %v", out.Transport.ServiceName, expectedService)
					}
				} else if tt.expectType != "" {
					if out.Transport == nil {
						t.Fatal("Transport config is nil")
					}
					if out.Transport.Type != tt.expectType {
						t.Errorf("Transport.Type = %v, want %v", out.Transport.Type, tt.expectType)
					}
					if tt.expectPath != "" && out.Transport.Path != tt.expectPath {
						t.Errorf("Transport.Path = %v, want %v", out.Transport.Path, tt.expectPath)
					}
					if len(tt.expectHosts) > 0 {
						hosts, ok := out.Transport.Host.([]string)
						if !ok {
							t.Errorf("Transport.Host type = %T, want []string", out.Transport.Host)
						} else if len(hosts) != len(tt.expectHosts) {
							t.Errorf("Transport.Host length = %v, want %v", len(hosts), len(tt.expectHosts))
						}
					}
				}
			})
		}
	})

	t.Run("特殊情况处理", func(t *testing.T) {
		t.Run("空节点", func(t *testing.T) {
			_, err := ConvertVmessToSingBox(nil)
			if err == nil {
				t.Error("Expected error for nil node, but got nil")
			}
		})

		t.Run("空标签", func(t *testing.T) {
			node := &VmessNode{
				Version: "2",
				Remarks: "", // 空备注
				Address: "example.com",
				Port:    "443",
				ID:      "test-uuid-1234",
				AlterID: "0",
				Network: "tcp",
			}

			out, err := ConvertVmessToSingBox(node)
			if err != nil {
				t.Fatalf("ConvertVmessToSingBox failed: %v", err)
			}

			if out.Tag != "vmess-node" { // 应该使用默认标签
				t.Errorf("Tag = %v, want %v", out.Tag, "vmess-node")
			}
		})

		t.Run("QUIC协议", func(t *testing.T) {
			node := &VmessNode{
				Version: "2",
				Remarks: "QUIC Node",
				Address: "example.com",
				Port:    "443",
				ID:      "test-uuid-1234",
				AlterID: "0",
				Network: "quic",
				TLS:     "tls",
			}

			out, err := ConvertVmessToSingBox(node)
			if err != nil {
				t.Fatalf("ConvertVmessToSingBox failed: %v", err)
			}

			// QUIC 不应该有 transport 配置
			if out.Transport != nil {
				t.Errorf("Transport should be nil for QUIC, but got %+v", out.Transport)
			}
		})
	})
}

// 辅助函数：base64 编码
func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
