package entity

import "encoding/json"

// SingBoxConfig 是最终返回给客户端的 sing-box 根配置。
// 这里的结构尽量贴近 sing-box 官方 JSON 字段，便于直接序列化输出。
type SingBoxConfig struct {
	DNS          SingDNS                 `json:"dns"`                    //https://sing-box.sagernet.org/zh/configuration/dns/
	Endpoints    []SingEndpointWireguard `json:"endpoints"`              //https://sing-box.sagernet.org/zh/configuration/endpoint/
	Route        SingRoute               `json:"route"`                  //https://sing-box.sagernet.org/zh/configuration/route/
	Experimental *SingExperimental       `json:"experimental,omitempty"` //https://sing-box.sagernet.org/zh/configuration/experimental/
	Inbounds     []SingInbound           `json:"inbounds"`               //https://sing-box.sagernet.org/zh/configuration/inbound/
	Outbounds    []SingBoxOut            `json:"outbounds"`
}

// SingBoxOut 统一承载多种 sing-box outbound 的公共字段。
// 因为不同协议字段差异较大，这里使用“大结构 + omitempty”的方式兼容多种配置形态。
type SingBoxOut struct {
	Username             string                    `json:"username,omitempty"`
	Password             string                    `json:"password,omitempty"`
	Server               string                    `json:"server,omitempty"`
	ServerPort           int                       `json:"server_port,omitempty"`
	Tag                  string                    `json:"tag,omitempty"`
	TLS                  *SingTLS                  `json:"tls,omitempty"`
	Transport            *SingTransport            `json:"transport,omitempty"`
	Type                 string                    `json:"type,omitempty"`
	Method               string                    `json:"method,omitempty"`
	AlterID              int                       `json:"alter_id,omitempty"`
	Security             string                    `json:"security,omitempty"`
	UUID                 string                    `json:"uuid,omitempty"`
	Default              string                    `json:"default,omitempty"`
	Outbounds            []string                  `json:"outbounds,omitempty"`
	Interval             string                    `json:"interval,omitempty"`
	Tolerance            int                       `json:"tolerance,omitempty"`
	URL                  string                    `json:"url,omitempty"`
	Network              string                    `json:"network,omitempty"`
	Plugin               string                    `json:"plugin,omitempty"`
	PluginOpts           string                    `json:"plugin_opts,omitempty"`
	ObfsParam            string                    `json:"obfs_param,omitempty"`
	Protocol             string                    `json:"protocol,omitempty"`
	ProtocolParam        string                    `json:"protocol_param,omitempty"`
	Flow                 string                    `json:"flow,omitempty"`
	PacketEncoding       string                    `json:"packet_encoding,omitempty"`
	AuthStr              string                    `json:"auth_str,omitempty"`
	DisableMtuDiscovery  bool                      `json:"disable_mtu_discovery,omitempty"`
	Down                 string                    `json:"down,omitempty"`
	DownMbps             int                       `json:"down_mbps,omitempty"`
	RecvWindow           int                       `json:"recv_window,omitempty"`
	RecvWindowConn       int                       `json:"recv_window_conn,omitempty"`
	Up                   string                    `json:"up,omitempty"`
	UpMbps               int                       `json:"up_mbps,omitempty"`
	Detour               string                    `json:"detour,omitempty"`
	Multiplex            *SingMultiplex            `json:"multiplex,omitempty"`
	Version              int                       `json:"version,omitempty"`
	UdpOverTcp           *SingUdpOverTcp           `json:"udp_over_tcp,omitempty"`
	SystemInterface      bool                      `json:"system_interface,omitempty"`
	InterfaceName        string                    `json:"interface_name,omitempty"`
	LocalAddress         []string                  `json:"local_address,omitempty"`
	PrivateKey           string                    `json:"private_key,omitempty"`
	Peers                []*SingWireguardMultiPeer `json:"peers,omitempty"`
	PeerPublicKey        string                    `json:"peer_public_key,omitempty"`
	PreSharedKey         string                    `json:"pre_shared_key,omitempty"`
	Reserved             []int64                   `json:"reserved,omitempty"`
	MTU                  uint                      `json:"mtu,omitempty"`
	CongestionController string                    `json:"congestion_control,omitempty"`
	UdpRelayMode         string                    `json:"udp_relay_mode,omitempty"`
	ZeroRttHandshake     bool                      `json:"zero_rtt_handshake,omitempty"`
	Heartbeat            string                    `json:"heartbeat,omitempty"`
	Obfs                 *SingObfs                 `json:"obfs,omitempty"`
	Ignored              bool                      `json:"-"`
	TcpFastOpen          bool                      `json:"tcp_fast_open,omitempty"`
	TcpMultiPath         bool                      `json:"tcp_multi_path,omitempty"`
	Visible              []string                  `json:"-"`
}

// SingUdpOverTcp 表示 sing-box 的 udp_over_tcp 开关。
type SingUdpOverTcp struct {
	Enabled bool `json:"enabled,omitempty"`
}

// SingTLS 描述 outbound 上的 TLS 配置。
type SingTLS struct {
	Enabled     bool         `json:"enabled,omitempty"`
	ServerName  string       `json:"server_name,omitempty"`
	Alpn        []string     `json:"alpn,omitempty"`
	Insecure    bool         `json:"insecure,omitempty"`
	Utls        *SingUtls    `json:"utls,omitempty"`
	Reality     *SingReality `json:"reality,omitempty"`
	Certificate string       `json:"certificate,omitempty"`
}

// SingUtls 表示 uTLS 指纹伪装配置。
type SingUtls struct {
	Enabled     bool   `json:"enabled,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// SingReality 表示 Reality 握手参数。
type SingReality struct {
	Enabled   bool   `json:"enabled,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	ShortID   string `json:"short_id,omitempty"`
}

// SingTransport 描述 ws / http / grpc 等传输层附加参数。
type SingTransport struct {
	Headers             map[string][]string `json:"headers,omitempty"`
	Path                string              `json:"path,omitempty"`
	Type                string              `json:"type,omitempty"`
	EarlyDataHeaderName string              `json:"early_data_header_name,omitempty"`
	MaxEarlyData        int                 `json:"max_early_data,omitempty"`
	Host                any                 `json:"host,omitempty"`
	Method              string              `json:"method,omitempty"`
	ServiceName         string              `json:"service_name,omitempty"`
}

// SingMultiplex 对应 sing-box 的多路复用配置。
type SingMultiplex struct {
	Enabled        bool   `json:"enabled,omitempty"`
	MaxConnections int    `json:"max_connections,omitempty"`
	MinStreams     int    `json:"min_streams,omitempty"`
	MaxStreams     int    `json:"max_streams,omitempty"`
	Padding        bool   `json:"padding,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
}

// SingWireguardMultiPeer 兼容 wireguard outbound 的多 peer 写法。
type SingWireguardMultiPeer struct {
	Server       string   `json:"server,omitempty"`
	ServerPort   int      `json:"server_port,omitempty"`
	PublicKey    string   `json:"public_key,omitempty"`
	PreSharedKey string   `json:"pre_shared_key,omitempty"`
	AllowedIps   []string `json:"allowed_ips,omitempty"`
	Reserved     []int64  `json:"reserved,omitempty"`
}

// SingObfs 用于兼容不同协议的混淆字段序列化格式。
// 当 Type 为空时直接输出字符串；否则输出 hysteria2 需要的对象结构。
type SingObfs struct {
	Value string
	Type  string
}

type singObfsHysteria2 struct {
	Password string `json:"password"`
	Type     string `json:"type"`
}

func (s SingObfs) MarshalJSON() ([]byte, error) {
	if s.Type == "" {
		return json.Marshal(s.Value)
	}
	ns := singObfsHysteria2{
		Password: s.Value,
		Type:     s.Type,
	}
	return json.Marshal(ns)
}

// SingDNS 是 sing-box DNS 配置根对象。
type SingDNS struct {
	Servers []SingDNSServer `json:"servers"`
	Rules   []SingDNSRule   `json:"rules"`
	Final   string          `json:"final"`
}

// SingDNSServer 表示一条 DNS 上游。
type SingDNSServer struct {
	Tag             string `json:"tag"`
	Address         string `json:"address"`
	AddressResolver string `json:"address_resolver,omitempty"`
	Strategy        string `json:"strategy,omitempty"`
	Detour          string `json:"detour,omitempty"`
}

// SingDNSRule 描述 DNS 分流规则。
type SingDNSRule struct {
	Outbound  string `json:"outbound,omitempty"`
	Server    string `json:"server,omitempty"`
	ClashMode string `json:"clash_mode,omitempty"`
	RuleSet   string `json:"rule_set,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Action    string `json:"action,omitempty"`
}

// SingEndpointWireguard 表示 sing-box endpoint.wireguard。
type SingEndpointWireguard struct {
	Type       string                      `json:"type"`
	Tag        string                      `json:"tag"`
	MTU        int                         `json:"mtu"`
	Address    []string                    `json:"address"`
	PrivateKey string                      `json:"private_key"`
	Peers      []SingEndpointWireguardPeer `json:"peers"`
}

// SingEndpointWireguardPeer 是 endpoint.wireguard 下的单个对端。
type SingEndpointWireguardPeer struct {
	Address                     string   `json:"address"`
	Port                        int      `json:"port"`
	PublicKey                   string   `json:"public_key"`
	PreSharedKey                string   `json:"pre_shared_key"`
	AllowedIps                  []string `json:"allowed_ips"`
	PersistentKeepaliveInterval int      `json:"persistent_keepalive_interval"`
}

// SingRoute 是 sing-box 路由配置根对象。
type SingRoute struct {
	Rules               []SingRouteRule `json:"rules"`
	RuleSet             []SingRuleSet   `json:"rule_set"`
	Final               string          `json:"final"`
	FindProcess         bool            `json:"find_process"`
	AutoDetectInterface bool            `json:"auto_detect_interface"`
}

// SingRouteRule 描述一条路由匹配规则。
type SingRouteRule struct {
	Protocol     string   `json:"protocol,omitempty"`
	Action       string   `json:"action,omitempty"`
	ClashMode    string   `json:"clash_mode,omitempty"`
	Outbound     string   `json:"outbound,omitempty"`
	Port         []int    `json:"port,omitempty"`
	IPIsPrivate  bool     `json:"ip_is_private,omitempty"`
	IPCIDR       []string `json:"ip_cidr,omitempty"`
	WifiSSID     []string `json:"wifi_ssid,omitempty"`
	DomainSuffix []string `json:"domain_suffix,omitempty"`
	RuleSet      []string `json:"rule_set,omitempty"`
}

// SingRuleSet 对应 sing-box route.rule_set 条目。
type SingRuleSet struct {
	Type           string      `json:"type"`
	Tag            string      `json:"tag"`
	Rules          interface{} `json:"rules,omitempty"`
	Format         string      `json:"format,omitempty"`
	URL            string      `json:"url,omitempty"`
	DownloadDetour string      `json:"download_detour,omitempty"`
}

// SingExperimental 汇总实验性配置。
type SingExperimental struct {
	CacheFile SingCacheFile `json:"cache_file"`
	ClashAPI  SingClashAPI  `json:"clash_api"`
}

// SingCacheFile 控制 cache_file 功能开关。
type SingCacheFile struct {
	Enabled bool `json:"enabled"`
}

// SingClashAPI 用来开启 clash_api 与面板 UI。
type SingClashAPI struct {
	ExternalController       string `json:"external_controller"`
	ExternalUI               string `json:"external_ui"`
	ExternalUIDownloadURL    string `json:"external_ui_download_url"`
	ExternalUIDownloadDetour string `json:"external_ui_download_detour"`
	DefaultMode              string `json:"default_mode"`
}

// SingInbound 是生成结果中的入站结构。
// 当前项目主要把管理端录入的 ConfigJSON 反序列化为该结构后输出。
type SingInbound struct {
	Type         string   `json:"type"`
	Tag          string   `json:"tag,omitempty"`
	Listen       string   `json:"listen,omitempty"`
	ListenPort   int      `json:"listen_port,omitempty"`
	Address      []string `json:"address,omitempty"`
	Inet4Address string   `json:"inet4_address,omitempty"`
	AutoRoute    bool     `json:"auto_route,omitempty"`
	Stack        string   `json:"stack,omitempty"`
	Sniff        bool     `json:"sniff,omitempty"`
}
