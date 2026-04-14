package singbox

import (
	"reflect"
	"singboxconfig/entity"
	"testing"
)

func TestResolveDNS(t *testing.T) {
	defaultDNS := GetDefaultDNS()

	if got := ResolveDNS(""); !reflect.DeepEqual(got, defaultDNS) {
		t.Fatalf("ResolveDNS(empty) = %+v, want %+v", got, defaultDNS)
	}

	if got := ResolveDNS("{"); !reflect.DeepEqual(got, defaultDNS) {
		t.Fatalf("ResolveDNS(invalid) = %+v, want %+v", got, defaultDNS)
	}

	custom := `{"servers":[{"tag":"custom","address":"1.1.1.1"}],"rules":[],"final":"custom"}`
	got := ResolveDNS(custom)
	if got.Final != "custom" || len(got.Servers) != 1 || got.Servers[0].Tag != "custom" {
		t.Fatalf("ResolveDNS(custom) got unexpected dns: %+v", got)
	}
}

func TestGetInbounds(t *testing.T) {
	items := []*entity.Inbound{
		{Tag: "bad", Enabled: true, Sort: 30, ConfigJSON: `{`},
		{Tag: "disabled", Enabled: false, Sort: 20, ConfigJSON: `{"type":"http"}`},
		{Tag: "tun", Enabled: true, Sort: 10, ConfigJSON: `{"type":"tun","stack":"mixed"}`},
		{Tag: "mixed", Enabled: true, Sort: 15, ConfigJSON: `{"type":"mixed","tag":"mixed-in"}`},
	}

	got := GetInbounds(items)
	if len(got) != 2 {
		t.Fatalf("GetInbounds len = %d, want 2", len(got))
	}
	if got[0].Type != "tun" || got[1].Type != "mixed" {
		t.Fatalf("GetInbounds order/types unexpected: %+v", got)
	}
}

func TestGetWireGuardEndpoints(t *testing.T) {
	device := &entity.Device{
		Code:                "phone",
		WireGuardTag:        "wg-default",
		WireGuardClientAddr: "10.8.0.4",
		WireGuardClientKey:  "client-key",
	}
	wg := &entity.WireGuard{
		Tag:         "wg-default",
		Enabled:     true,
		EndpointTag: "wg-ep",
		MTU:         1408,
	}
	peers := []*entity.WireGuardPeer{
		{ID: 2, Sort: 20, WireGuardTag: "wg-default", Address: "2.2.2.2", Port: 2, PublicKey: "pk2", AllowedIPs: "10.0.0.0/24", Enabled: true},
		{ID: 1, Sort: 10, WireGuardTag: "wg-default", Address: "1.1.1.1", Port: 1, PublicKey: "pk1", AllowedIPs: "10.8.0.0/24, 192.168.10.0/24", Enabled: true},
		{ID: 3, Sort: 30, WireGuardTag: "wg-default", Address: "3.3.3.3", Port: 3, PublicKey: "pk3", AllowedIPs: "0.0.0.0/0", Enabled: false},
	}

	got := GetWireGuardEndpoints(wg, peers, device)
	if len(got) != 1 {
		t.Fatalf("GetWireGuardEndpoints len = %d, want 1", len(got))
	}
	if got[0].Address[0] != "10.8.0.4/32" {
		t.Fatalf("GetWireGuardEndpoints address = %v, want 10.8.0.4/32", got[0].Address)
	}
	if len(got[0].Peers) != 2 || got[0].Peers[0].Address != "1.1.1.1" {
		t.Fatalf("GetWireGuardEndpoints peers unexpected: %+v", got[0].Peers)
	}
	if !reflect.DeepEqual(got[0].Peers[0].AllowedIps, []string{"10.8.0.0/24", "192.168.10.0/24"}) {
		t.Fatalf("GetWireGuardEndpoints allowed_ips unexpected: %+v", got[0].Peers[0].AllowedIps)
	}
}

func TestGetExtraOutbounds(t *testing.T) {
	items := []*entity.Outbound{
		{Tag: "bad", Enabled: true, Sort: 30, VisibleDevices: "phone", ConfigJSON: `{`},
		{Tag: "hidden", Enabled: true, Sort: 20, VisibleDevices: "tv", ConfigJSON: `{"tag":"hidden","type":"socks"}`},
		{Tag: "disabled", Enabled: false, Sort: 10, ConfigJSON: `{"tag":"disabled","type":"socks"}`},
		{Tag: "visible", Enabled: true, Sort: 5, VisibleDevices: "phone, office", ConfigJSON: `{"tag":"visible","type":"socks"}`},
	}

	got := GetExtraOutbounds("phone", items)
	if len(got) != 1 || got[0].Tag != "visible" {
		t.Fatalf("GetExtraOutbounds unexpected result: %+v", got)
	}
}
