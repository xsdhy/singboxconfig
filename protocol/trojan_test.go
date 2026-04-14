package protocol

import (
	"testing"
)

func Test_DecodeTrojanUrl(t *testing.T) {
	inputURL := "trojan://1f109a88-3660-4214-bce2-2ed4d1e58316@lbso.bnnodeservice.com:32443?security=tls&sni=cert.bitbyte.one&allowInsecure=1&peer=cert.bitbyte.one&type=tcp&host=cert.bitbyte.one#%F0%9F%87%AD%F0%9F%87%B0%20%E9%A6%99%E6%B8%AF-%E5%B9%BF%E4%B8%9C%E4%B8%93%E7%BA%BF%20BGP%201"
	expectationItem := &TrojanNode{
		Scheme:        "trojan",
		Password:      "1f109a88-3660-4214-bce2-2ed4d1e58316",
		Host:          "lbso.bnnodeservice.com",
		Port:          32443,
		Security:      "tls",
		SNI:           "cert.bitbyte.one",
		AllowInsecure: true,
		Peer:          "cert.bitbyte.one",
		Type:          "tcp",
		HostHeader:    "cert.bitbyte.one",
		Tag:           "香港广东专线BGP1",
	}

	item, err := DecodeTrojanUrl(inputURL)
	if err != nil {
		t.Fatalf("DecodeTrojanUrl failed: %v", err)
	}

	// 验证所有字段
	if item.Scheme != expectationItem.Scheme {
		t.Errorf("Scheme = %v, want %v", item.Scheme, expectationItem.Scheme)
	}
	if item.Password != expectationItem.Password {
		t.Errorf("Password = %v, want %v", item.Password, expectationItem.Password)
	}
	if item.Host != expectationItem.Host {
		t.Errorf("Host = %v, want %v", item.Host, expectationItem.Host)
	}
	if item.Port != expectationItem.Port {
		t.Errorf("Port = %v, want %v", item.Port, expectationItem.Port)
	}
	if item.Security != expectationItem.Security {
		t.Errorf("Security = %v, want %v", item.Security, expectationItem.Security)
	}
	if item.SNI != expectationItem.SNI {
		t.Errorf("SNI = %v, want %v", item.SNI, expectationItem.SNI)
	}
	if item.AllowInsecure != expectationItem.AllowInsecure {
		t.Errorf("AllowInsecure = %v, want %v", item.AllowInsecure, expectationItem.AllowInsecure)
	}
	if item.Peer != expectationItem.Peer {
		t.Errorf("Peer = %v, want %v", item.Peer, expectationItem.Peer)
	}
	if item.Type != expectationItem.Type {
		t.Errorf("Type = %v, want %v", item.Type, expectationItem.Type)
	}
	if item.HostHeader != expectationItem.HostHeader {
		t.Errorf("HostHeader = %v, want %v", item.HostHeader, expectationItem.HostHeader)
	}
	if item.Tag != expectationItem.Tag {
		t.Errorf("Tag = %v, want %v", item.Tag, expectationItem.Tag)
	}
}

func Test_ConvertTrojanToSingBox(t *testing.T) {
	node := &TrojanNode{
		Scheme:        "trojan",
		Password:      "1f109a88-3660-4214-bce2-2ed4d1e58316",
		Host:          "lbso.bnnodeservice.com",
		Port:          32443,
		Security:      "tls",
		SNI:           "cert.bitbyte.one",
		AllowInsecure: true,
		Peer:          "cert.bitbyte.one",
		Type:          "tcp",
		HostHeader:    "cert.bitbyte.one",
		Tag:           "🇭🇰 香港-广东专线 BGP 1",
	}

	out, err := ConvertTrojanToSingBox(node)
	if err != nil {
		t.Fatalf("ConvertTrojanToSingBox failed: %v", err)
	}

	// 验证基本字段
	if out.Type != "trojan" {
		t.Errorf("Type = %v, want %v", out.Type, "trojan")
	}
	if out.Tag != node.Tag {
		t.Errorf("Tag = %v, want %v", out.Tag, node.Tag)
	}
	if out.Server != node.Host {
		t.Errorf("Server = %v, want %v", out.Server, node.Host)
	}
	if out.ServerPort != node.Port {
		t.Errorf("ServerPort = %v, want %v", out.ServerPort, node.Port)
	}
	if out.Password != node.Password {
		t.Errorf("Password = %v, want %v", out.Password, node.Password)
	}
	if out.Network != node.Type {
		t.Errorf("Network = %v, want %v", out.Network, node.Type)
	}

	// 验证 TLS 配置
	if out.TLS == nil {
		t.Fatal("TLS config is nil")
	}
	if out.TLS.Enabled != (node.Security == "tls") {
		t.Errorf("TLS.Enabled = %v, want %v", out.TLS.Enabled, node.Security == "tls")
	}
	if out.TLS.ServerName != node.Peer {
		t.Errorf("TLS.ServerName = %v, want %v", out.TLS.ServerName, node.Peer)
	}
	if out.TLS.Insecure != node.AllowInsecure {
		t.Errorf("TLS.Insecure = %v, want %v", out.TLS.Insecure, node.AllowInsecure)
	}
}

func Test_DecodeTrojanUrl_SecondCase(t *testing.T) {
	inputURL := "trojan://51a4fec0-79ec-3e30-ae5e-5c2f882e9cf4@hk.pro.yypa.zzssptop.com:20475?peer=hkt1.zzsspnode.com#%E3%80%90%E5%AE%98%20%E7%BD%91%E3%80%91yywl.org"
	expectationItem := &TrojanNode{
		Scheme:        "trojan",
		Password:      "51a4fec0-79ec-3e30-ae5e-5c2f882e9cf4",
		Host:          "hk.pro.yypa.zzssptop.com",
		Port:          20475,
		Security:      "",
		SNI:           "",
		AllowInsecure: false,
		Peer:          "hkt1.zzsspnode.com",
		Type:          "",
		HostHeader:    "",
		Tag:           "官网yywlorg",
	}

	item, err := DecodeTrojanUrl(inputURL)
	if err != nil {
		t.Fatalf("DecodeTrojanUrl failed: %v", err)
	}

	// 验证所有字段
	if item.Scheme != expectationItem.Scheme {
		t.Errorf("Scheme = %v, want %v", item.Scheme, expectationItem.Scheme)
	}
	if item.Password != expectationItem.Password {
		t.Errorf("Password = %v, want %v", item.Password, expectationItem.Password)
	}
	if item.Host != expectationItem.Host {
		t.Errorf("Host = %v, want %v", item.Host, expectationItem.Host)
	}
	if item.Port != expectationItem.Port {
		t.Errorf("Port = %v, want %v", item.Port, expectationItem.Port)
	}
	if item.Security != expectationItem.Security {
		t.Errorf("Security = %v, want %v", item.Security, expectationItem.Security)
	}
	if item.SNI != expectationItem.SNI {
		t.Errorf("SNI = %v, want %v", item.SNI, expectationItem.SNI)
	}
	if item.AllowInsecure != expectationItem.AllowInsecure {
		t.Errorf("AllowInsecure = %v, want %v", item.AllowInsecure, expectationItem.AllowInsecure)
	}
	if item.Peer != expectationItem.Peer {
		t.Errorf("Peer = %v, want %v", item.Peer, expectationItem.Peer)
	}
	if item.Type != expectationItem.Type {
		t.Errorf("Type = %v, want %v", item.Type, expectationItem.Type)
	}
	if item.HostHeader != expectationItem.HostHeader {
		t.Errorf("HostHeader = %v, want %v", item.HostHeader, expectationItem.HostHeader)
	}
	if item.Tag != expectationItem.Tag {
		t.Errorf("Tag = %v, want %v", item.Tag, expectationItem.Tag)
	}
}

func Test_DecodeTrojanUrl_ThirdCase(t *testing.T) {
	inputURL := "trojan://51a4fec0-79ec-3e30-ae5e-5c2f882e9cf4@9d8a3e.old.yyp.zzssptop.com:20500?peer=sg3.zzsspnode.com#A%E6%96%B0%E5%8A%A0%E5%9D%A103"
	expectationItem := &TrojanNode{
		Scheme:        "trojan",
		Password:      "51a4fec0-79ec-3e30-ae5e-5c2f882e9cf4",
		Host:          "9d8a3e.old.yyp.zzssptop.com",
		Port:          20500,
		Security:      "",
		SNI:           "",
		AllowInsecure: false,
		Peer:          "sg3.zzsspnode.com",
		Type:          "",
		HostHeader:    "",
		Tag:           "A新加坡03",
	}

	item, err := DecodeTrojanUrl(inputURL)
	if err != nil {
		t.Fatalf("DecodeTrojanUrl failed: %v", err)
	}

	// 验证所有字段
	if item.Scheme != expectationItem.Scheme {
		t.Errorf("Scheme = %v, want %v", item.Scheme, expectationItem.Scheme)
	}
	if item.Password != expectationItem.Password {
		t.Errorf("Password = %v, want %v", item.Password, expectationItem.Password)
	}
	if item.Host != expectationItem.Host {
		t.Errorf("Host = %v, want %v", item.Host, expectationItem.Host)
	}
	if item.Port != expectationItem.Port {
		t.Errorf("Port = %v, want %v", item.Port, expectationItem.Port)
	}
	if item.Security != expectationItem.Security {
		t.Errorf("Security = %v, want %v", item.Security, expectationItem.Security)
	}
	if item.SNI != expectationItem.SNI {
		t.Errorf("SNI = %v, want %v", item.SNI, expectationItem.SNI)
	}
	if item.AllowInsecure != expectationItem.AllowInsecure {
		t.Errorf("AllowInsecure = %v, want %v", item.AllowInsecure, expectationItem.AllowInsecure)
	}
	if item.Peer != expectationItem.Peer {
		t.Errorf("Peer = %v, want %v", item.Peer, expectationItem.Peer)
	}
	if item.Type != expectationItem.Type {
		t.Errorf("Type = %v, want %v", item.Type, expectationItem.Type)
	}
	if item.HostHeader != expectationItem.HostHeader {
		t.Errorf("HostHeader = %v, want %v", item.HostHeader, expectationItem.HostHeader)
	}
	if item.Tag != expectationItem.Tag {
		t.Errorf("Tag = %v, want %v", item.Tag, expectationItem.Tag)
	}
}
