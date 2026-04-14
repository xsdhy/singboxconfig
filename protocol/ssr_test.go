package protocol

import (
	"testing"
)

func TestDecodeSSRURL(t *testing.T) {
	// 测试直接格式（非base64编码的整个URL）
	t.Run("直接格式URL", func(t *testing.T) {
		testURL := "ssr://9d8a3e.old.yyp.zzssptop.com:20572:auth_aes128_sha1:chacha20-ietf:plain:aDBHRVdP/?obfsparam=YjUxNzMzNDM4Mi5kb3dubG9hZC53aW5kb3dzdXBkYXRlLmNvbQ&protoparam=MzQzODI6ZDhxZmU0&remarks=QeS_hOe9l-aWrzAx"

		node, err := DecodeSSRURL(testURL)
		if err != nil {
			t.Fatalf("DecodeSSRURL failed: %v", err)
		}

		// 验证基本字段
		if node.Server != "9d8a3e.old.yyp.zzssptop.com" {
			t.Errorf("Expected server '9d8a3e.old.yyp.zzssptop.com', got '%s'", node.Server)
		}

		if node.Port != 20572 {
			t.Errorf("Expected port 20572, got %d", node.Port)
		}

		if node.Protocol != "auth_aes128_sha1" {
			t.Errorf("Expected protocol 'auth_aes128_sha1', got '%s'", node.Protocol)
		}

		if node.Method != "chacha20-ietf" {
			t.Errorf("Expected method 'chacha20-ietf', got '%s'", node.Method)
		}

		if node.Obfs != "plain" {
			t.Errorf("Expected obfs 'plain', got '%s'", node.Obfs)
		}

		// 验证密码解码
		t.Logf("Password: %s", node.Password)

		// 验证参数解码
		t.Logf("ObfsParam: %s", node.ObfsParam)
		t.Logf("ProtocolParam: %s", node.ProtocolParam)
		t.Logf("Remarks: %s", node.Remarks)
		t.Logf("Tag: %s", node.Tag)
	})

	// 测试Base64编码的整个URL格式
	t.Run("Base64编码的整个URL", func(t *testing.T) {
		testURL := "ssr://OWQ4YTNlLm9sZC55eXAuenpzc3B0b3AuY29tOjIwNTcyOmF1dGhfYWVzMTI4X3NoYTE6Y2hhY2hhMjAtaWV0ZjpwbGFpbjphREJIUlZkUC8_b2Jmc3BhcmFtPVlqVXhOek16TkRNNE1pNWtiM2R1Ykc5aFpDNTNhVzVrYjNkemRYQmtZWFJsTG1OdmJRJnByb3RvcGFyYW09TXpRek9ESTZaRGh4Wm1VMCZyZW1hcmtzPVFlU19oT2U5bC1hV3J6QXg"

		node, err := DecodeSSRURL(testURL)
		if err != nil {
			t.Fatalf("DecodeSSRURL failed: %v", err)
		}

		// 验证基本字段
		if node.Server != "9d8a3e.old.yyp.zzssptop.com" {
			t.Errorf("Expected server '9d8a3e.old.yyp.zzssptop.com', got '%s'", node.Server)
		}

		if node.Port != 20572 {
			t.Errorf("Expected port 20572, got %d", node.Port)
		}

		if node.Protocol != "auth_aes128_sha1" {
			t.Errorf("Expected protocol 'auth_aes128_sha1', got '%s'", node.Protocol)
		}

		if node.Method != "chacha20-ietf" {
			t.Errorf("Expected method 'chacha20-ietf', got '%s'", node.Method)
		}

		if node.Obfs != "plain" {
			t.Errorf("Expected obfs 'plain', got '%s'", node.Obfs)
		}

		// 验证密码解码
		expectedPassword := "h0GEWO"
		if node.Password != expectedPassword {
			t.Errorf("Expected password '%s', got '%s'", expectedPassword, node.Password)
		}

		// 验证参数解码
		expectedObfsParam := "b517334382.download.windowsupdate.com"
		if node.ObfsParam != expectedObfsParam {
			t.Errorf("Expected obfs_param '%s', got '%s'", expectedObfsParam, node.ObfsParam)
		}

		expectedProtocolParam := "34382:d8qfe4"
		if node.ProtocolParam != expectedProtocolParam {
			t.Errorf("Expected protocol_param '%s', got '%s'", expectedProtocolParam, node.ProtocolParam)
		}

		expectedRemarks := "A俄罗斯01"
		if node.Remarks != expectedRemarks {
			t.Errorf("Expected remarks '%s', got '%s'", expectedRemarks, node.Remarks)
		}

		t.Logf("Password: %s", node.Password)
		t.Logf("ObfsParam: %s", node.ObfsParam)
		t.Logf("ProtocolParam: %s", node.ProtocolParam)
		t.Logf("Remarks: %s", node.Remarks)
		t.Logf("Tag: %s", node.Tag)
	})
}

func TestConvertSSRToSingBox(t *testing.T) {
	node := &SSRNode{
		Server:        "example.com",
		Port:          443,
		Protocol:      "auth_aes128_sha1",
		Method:        "chacha20-ietf",
		Obfs:          "plain",
		Password:      "password123",
		ObfsParam:     "obfs.example.com",
		ProtocolParam: "12345:abcdef",
		Remarks:       "Test Node",
		Tag:           "Test Node",
	}

	singBox, err := ConvertSSRToSingBox(node)
	if err != nil {
		t.Fatalf("ConvertSSRToSingBox failed: %v", err)
	}

	if singBox.Type != "shadowsocksr" {
		t.Errorf("Expected type 'shadowsocksr', got '%s'", singBox.Type)
	}

	if singBox.Server != node.Server {
		t.Errorf("Expected server '%s', got '%s'", node.Server, singBox.Server)
	}

	if singBox.ServerPort != node.Port {
		t.Errorf("Expected port %d, got %d", node.Port, singBox.ServerPort)
	}

	if singBox.Method != node.Method {
		t.Errorf("Expected method '%s', got '%s'", node.Method, singBox.Method)
	}

	if singBox.Password != node.Password {
		t.Errorf("Expected password '%s', got '%s'", node.Password, singBox.Password)
	}

	if singBox.Protocol != node.Protocol {
		t.Errorf("Expected protocol '%s', got '%s'", node.Protocol, singBox.Protocol)
	}

	if singBox.ProtocolParam != node.ProtocolParam {
		t.Errorf("Expected protocol_param '%s', got '%s'", node.ProtocolParam, singBox.ProtocolParam)
	}

	if singBox.Obfs == nil || singBox.Obfs.Value != node.Obfs {
		obfsValue := ""
		if singBox.Obfs != nil {
			obfsValue = singBox.Obfs.Value
		}
		t.Errorf("Expected obfs '%s', got '%s'", node.Obfs, obfsValue)
	}

	if singBox.ObfsParam != node.ObfsParam {
		t.Errorf("Expected obfs_param '%s', got '%s'", node.ObfsParam, singBox.ObfsParam)
	}
}

func TestDecodeSSRURLToSingBox(t *testing.T) {
	// 测试直接格式
	t.Run("直接格式转换", func(t *testing.T) {
		testURL := "ssr://9d8a3e.old.yyp.zzssptop.com:20572:auth_aes128_sha1:chacha20-ietf:plain:aDBHRVdP/?obfsparam=YjUxNzMzNDM4Mi5kb3dubG9hZC53aW5kb3dzdXBkYXRlLmNvbQ&protoparam=MzQzODI6ZDhxZmU0&remarks=QeS_hOe9l-aWrzAx"

		singBox, err := DecodeSSRURLToSingBox(testURL)
		if err != nil {
			t.Fatalf("DecodeSSRURLToSingBox failed: %v", err)
		}

		if singBox.Type != "shadowsocksr" {
			t.Errorf("Expected type 'shadowsocksr', got '%s'", singBox.Type)
		}

		t.Logf("SingBox config: %+v", singBox)
	})

	// 测试Base64编码格式
	t.Run("Base64编码格式转换", func(t *testing.T) {
		testURL := "ssr://OWQ4YTNlLm9sZC55eXAuenpzc3B0b3AuY29tOjIwNTcyOmF1dGhfYWVzMTI4X3NoYTE6Y2hhY2hhMjAtaWV0ZjpwbGFpbjphREJIUlZkUC8_b2Jmc3BhcmFtPVlqVXhOek16TkRNNE1pNWtiM2R1Ykc5aFpDNTNhVzVrYjNkemRYQmtZWFJsTG1OdmJRJnByb3RvcGFyYW09TXpRek9ESTZaRGh4Wm1VMCZyZW1hcmtzPVFlU19oT2U5bC1hV3J6QXg"

		singBox, err := DecodeSSRURLToSingBox(testURL)
		if err != nil {
			t.Fatalf("DecodeSSRURLToSingBox failed: %v", err)
		}

		if singBox.Type != "shadowsocksr" {
			t.Errorf("Expected type 'shadowsocksr', got '%s'", singBox.Type)
		}

		if singBox.Server != "9d8a3e.old.yyp.zzssptop.com" {
			t.Errorf("Expected server '9d8a3e.old.yyp.zzssptop.com', got '%s'", singBox.Server)
		}

		if singBox.ServerPort != 20572 {
			t.Errorf("Expected port 20572, got %d", singBox.ServerPort)
		}

		t.Logf("SingBox config: %+v", singBox)
	})
}

func TestDecodeBase64WithPadding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "标准base64",
			input:    "aGVsbG8=",
			expected: "hello",
		},
		{
			name:     "缺少padding的base64",
			input:    "aGVsbG8",
			expected: "hello",
		},
		{
			name:     "URL安全的base64",
			input:    "aDBHRVdP",
			expected: "h0GEWO",
		},
		{
			name:     "非base64字符串",
			input:    "plaintext",
			expected: "plaintext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := decodeBase64WithPadding(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
