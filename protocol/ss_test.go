package protocol

import (
	"testing"
)

func Test_DecodeSSURL(t *testing.T) {
	ssUrl := "ss://YWVzLTEyOC1nY206ZDhxZmU0@dns.yypa.zzssptop.com:20475/?group=5LqR57-8572R57uc#AA%E9%A6%99%E6%B8%AF01"
	expected := &SSNode{
		Method:   "aes-128-gcm",
		Password: "d8qfe4",
		Host:     "dns.yypa.zzssptop.com",
		Port:     20475,
		Group:    "5LqR57-8572R57uc",
		Tag:      "AA香港01",
	}

	ssInfo, err := DecodeSSURL(ssUrl)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if ssInfo.Method != expected.Method {
		t.Errorf("Expected Method %s, but got %s", expected.Method, ssInfo.Method)
	}
	if ssInfo.Password != expected.Password {
		t.Errorf("Expected Password %s, but got %s", expected.Password, ssInfo.Password)
	}
	if ssInfo.Host != expected.Host {
		t.Errorf("Expected Host %s, but got %s", expected.Host, ssInfo.Host)
	}
	if ssInfo.Port != expected.Port {
		t.Errorf("Expected Port %d, but got %d", expected.Port, ssInfo.Port)
	}
	if ssInfo.Group != expected.Group {
		t.Errorf("Expected Group %s, but got %s", expected.Group, ssInfo.Group)
	}
	if ssInfo.Tag != expected.Tag {
		t.Errorf("Expected Tag %s, but got %s", expected.Tag, ssInfo.Tag)
	}
}

func Test_DecodeSSURL_WithPlugin(t *testing.T) {
	ssUrl := "ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTp0MHNybWR4cm0zeHlqbnZxejlld2x4YjJteXE3cmp1dg==@a583855.r8.glados-config.com:2377?plugin=obfs-local%3Bobfs%3Dtls%3Bobfs-host%3Da583855.default.mozilla.net%3A179358&uot=1#Fast-B1-1"
	expected := &SSNode{
		Method:   "chacha20-ietf-poly1305",
		Password: "t0srmdxrm3xyjnvqz9ewlxb2myq7rjuv",
		Host:     "a583855.r8.glados-config.com",
		Port:     2377,
		Tag:      "Fast-B1-1",
	}

	ssInfo, err := DecodeSSURL(ssUrl)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if ssInfo.Method != expected.Method {
		t.Errorf("Expected Method %s, but got %s", expected.Method, ssInfo.Method)
	}
	if ssInfo.Password != expected.Password {
		t.Errorf("Expected Password %s, but got %s", expected.Password, ssInfo.Password)
	}
	if ssInfo.Host != expected.Host {
		t.Errorf("Expected Host %s, but got %s", expected.Host, ssInfo.Host)
	}
	if ssInfo.Port != expected.Port {
		t.Errorf("Expected Port %d, but got %d", expected.Port, ssInfo.Port)
	}
	if ssInfo.Tag != expected.Tag {
		t.Errorf("Expected Tag %s, but got %s", expected.Tag, ssInfo.Tag)
	}
}
