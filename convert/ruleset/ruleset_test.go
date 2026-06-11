package ruleset

import (
	"encoding/json"
	"errors"
	"reflect"
	"singboxconfig/entity"
	"testing"
)

// TestNormalizeSingboxAcceptsThreeForms 验证三种输入形态都能规范化为 {"version":1,"rules":[...]}，
// 且原始规则字段无损保留。
func TestNormalizeSingboxAcceptsThreeForms(t *testing.T) {
	cases := map[string]string{
		"full":   `{"version":1,"rules":[{"domain_suffix":["example.com"]}]}`,
		"array":  `[{"domain_suffix":["example.com"]}]`,
		"single": `{"domain_suffix":["example.com"]}`,
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := NormalizeSingbox(content)
			if err != nil {
				t.Fatalf("NormalizeSingbox error: %v", err)
			}
			var parsed singboxSourceRuleSet
			if err := json.Unmarshal(got, &parsed); err != nil {
				t.Fatalf("output not valid json: %v", err)
			}
			if parsed.Version != 1 {
				t.Errorf("version = %d, want 1", parsed.Version)
			}
			if len(parsed.Rules) != 1 {
				t.Fatalf("rules len = %d, want 1", len(parsed.Rules))
			}
			var rule map[string][]string
			if err := json.Unmarshal(parsed.Rules[0], &rule); err != nil {
				t.Fatalf("rule not object: %v", err)
			}
			if !reflect.DeepEqual(rule["domain_suffix"], []string{"example.com"}) {
				t.Errorf("domain_suffix = %v, want [example.com]", rule["domain_suffix"])
			}
		})
	}
}

// TestNormalizeSingboxInvalidContent 验证非法 Content 返回错误，调用方据此返回 500 或回退。
func TestNormalizeSingboxInvalidContent(t *testing.T) {
	for _, content := range []string{"", "   ", "not-json", "123", `"a string"`} {
		if _, err := NormalizeSingbox(content); err == nil {
			t.Errorf("content %q expected error, got nil", content)
		}
	}
}

// TestRenderLinesFieldMapping 验证各字段按固定顺序渲染，IPv4/IPv6 区分正确，geoip 转大写。
func TestRenderLinesFieldMapping(t *testing.T) {
	content := `{"version":1,"rules":[{
		"domain":["a.com"],
		"domain_suffix":["b.com"],
		"domain_keyword":["key"],
		"domain_regex":["^c.*"],
		"ip_cidr":["1.2.3.0/24","2001:db8::/32"],
		"geoip":["cn"],
		"process_name":["WeChat"],
		"process_path":["/Applications/QoderWork CN.app/Contents/MacOS/QoderWork CN","/Applications/WeChat.app/Contents/MacOS/WeChat"]
	}]}`
	lines, warnings, err := RenderLines(content)
	if err != nil {
		t.Fatalf("RenderLines error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	want := []string{
		"DOMAIN,a.com",
		"DOMAIN-SUFFIX,b.com",
		"DOMAIN-KEYWORD,key",
		"DOMAIN-REGEX,^c.*",
		"IP-CIDR,1.2.3.0/24",
		"IP-CIDR6,2001:db8::/32",
		"GEOIP,CN,no-resolve",
		"PROCESS-NAME,WeChat",
		"PROCESS-NAME,\"/Applications/QoderWork CN.app/Contents/MacOS/QoderWork CN\"",
		"PROCESS-NAME,\"/Applications/WeChat.app/Contents/MacOS/WeChat\"",
	}
	if !reflect.DeepEqual(lines, want) {
		t.Errorf("lines = %#v\nwant %#v", lines, want)
	}
}

// TestRenderLinesUnsupportedFieldWarns 验证未支持字段被跳过并记录 warning，不中断其它规则输出。
func TestRenderLinesUnsupportedFieldWarns(t *testing.T) {
	content := `[{"domain":["a.com"],"port":[80]}]`
	lines, warnings, err := RenderLines(content)
	if err != nil {
		t.Fatalf("RenderLines error: %v", err)
	}
	if !reflect.DeepEqual(lines, []string{"DOMAIN,a.com"}) {
		t.Errorf("lines = %v, want [DOMAIN,a.com]", lines)
	}
	if len(warnings) == 0 {
		t.Errorf("expected warning for unsupported field port")
	}
}

// TestRenderLinesInvalidContent 验证非法 Content 返回 ErrInvalidContent。
func TestRenderLinesInvalidContent(t *testing.T) {
	if _, _, err := RenderLines("not-json"); !errors.Is(err, ErrInvalidContent) && err == nil {
		t.Errorf("expected error for invalid content")
	}
	if _, _, err := RenderLines(""); err == nil {
		t.Errorf("expected error for empty content")
	}
}

// TestBuildRuleSetURL 验证 URL 拼接对特殊字符正确转义，路径只保留 tag，software / device / token 走 query 参数。
func TestBuildRuleSetURL(t *testing.T) {
	got := BuildRuleSetURL("https://config.example.com/", "geosite cn", entity.SoftwareSingbox, "iphone/15", "a?b c")
	want := "https://config.example.com/open/rules/geosite%20cn?device=iphone%2F15&software=singbox&token=a%3Fb+c"
	if got != want {
		t.Errorf("BuildRuleSetURL = %q\nwant %q", got, want)
	}
}

// TestCountLines 验证规则条数按展开后的逐行规则计数，且内容非法时返回 error。
func TestCountLines(t *testing.T) {
	// 1 个规则对象、2 个域名后缀 → 2 条。
	count, err := CountLines(`{"version":1,"rules":[{"domain_suffix":["a.com","b.com"]}]}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	if _, err := CountLines("not-json"); err == nil {
		t.Errorf("expected error for invalid content")
	}
}
