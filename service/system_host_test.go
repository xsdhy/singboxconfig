package service

import "testing"

// TestValidateSystemHost 覆盖系统 Host 校验的合法与非法分支。
func TestValidateSystemHost(t *testing.T) {
	valid := []string{
		"",                            // 空串表示未配置，允许
		"  ",                          // 纯空白等价于未配置
		"https://config.example.com",  // 标准 https
		"http://10.0.0.1:7391",        // http + 端口
		"https://config.example.com/", // 带尾斜杠，规范化后仍合法
	}
	for _, host := range valid {
		if err := validateSystemHost(host); err != nil {
			t.Errorf("validateSystemHost(%q) unexpected error: %v", host, err)
		}
	}

	invalid := []string{
		"config.example.com",       // 缺少 scheme
		"ftp://config.example.com", // 非 http/https
		"https://",                 // 缺少主机名
		"://nohost",                // 非法 URL
	}
	for _, host := range invalid {
		if err := validateSystemHost(host); err == nil {
			t.Errorf("validateSystemHost(%q) expected error, got nil", host)
		}
	}
}

// TestNormalizeSystemHost 验证去掉首尾空白与尾部斜杠。
func TestNormalizeSystemHost(t *testing.T) {
	cases := map[string]string{
		"  https://h.com/  ": "https://h.com",
		"https://h.com///":   "https://h.com",
		"https://h.com":      "https://h.com",
		"":                   "",
	}
	for in, want := range cases {
		if got := normalizeSystemHost(in); got != want {
			t.Errorf("normalizeSystemHost(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestNormalizeGlobalSettingValue 验证仅 system_host 走校验，其它 key 原样返回。
func TestNormalizeGlobalSettingValue(t *testing.T) {
	// system_host：规范化后返回去尾斜杠的值
	got, err := normalizeGlobalSettingValue(systemHostSettingKey, "https://h.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://h.com" {
		t.Errorf("got %q, want https://h.com", got)
	}

	// system_host 非法值：返回错误
	if _, err := normalizeGlobalSettingValue(systemHostSettingKey, "not-a-url"); err == nil {
		t.Errorf("expected error for invalid system_host")
	}

	// 其它 key：原样返回，不做校验
	got, err = normalizeGlobalSettingValue("dns_config", "{not normalized}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "{not normalized}" {
		t.Errorf("got %q, want unchanged", got)
	}
}
