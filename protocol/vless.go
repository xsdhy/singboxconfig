package protocol

import (
	"fmt"
	"net/url"
	"singboxconfig/entity"
	"strconv"
)

// VlessNode 表示从 VLESS URL 解析出的节点信息。
// VLESS 是 Xray 提出的无状态轻量代理协议，相比 VMess 去除了内置加密（加密由外层 TLS/Reality 承担）。
// URL 格式：vless://<uuid>@<host>:<port>?<params>#<tag>
type VlessNode struct {
	UUID        string `json:"uuid"`         // 用户 UUID，用于身份认证
	Host        string `json:"host"`         // 服务器地址（域名或 IP）
	Port        int    `json:"port"`         // 服务器端口号
	Security    string `json:"security"`     // 安全层类型：reality / tls / none（或空）
	Encryption  string `json:"encryption"`   // 加密方式，VLESS 固定为 none
	Type        string `json:"type"`         // 传输协议类型：tcp / ws / grpc / httpupgrade
	Flow        string `json:"flow"`         // 流控方式，Reality 场景下常用 xtls-rprx-vision
	PublicKey   string `json:"public_key"`   // Reality 服务器公钥（参数名 pbk）
	ShortID     string `json:"short_id"`     // Reality ShortID（参数名 sid），用于区分不同客户端
	SNI         string `json:"sni"`          // TLS/Reality SNI（Server Name Indication），优先取 sni 参数，其次取 servername
	Fingerprint string `json:"fingerprint"`  // uTLS 指纹伪装，如 chrome / firefox / safari（参数名 fp）
	Path        string `json:"path"`         // 传输路径（WebSocket / HTTPUpgrade 使用）
	HostHeader  string `json:"host_header"`  // 传输层 Host 头（WebSocket 使用，参数名 host）
	ServiceName string `json:"service_name"` // gRPC ServiceName（参数名 serviceName）
	Tag         string `json:"tag"`          // 节点备注标签（URL fragment 解码后清洗）
}

// DecodeVlessUrl 解析 VLESS URL，返回结构化的 VlessNode。
// 参考：https://github.com/XTLS/Xray-core/discussions/716
func DecodeVlessUrl(urlStr string) (*VlessNode, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	// 验证协议头
	if parsedURL.Scheme != "vless" {
		return nil, fmt.Errorf("invalid scheme: %s", parsedURL.Scheme)
	}

	// UUID 位于 URL userinfo 部分（@ 符号前）
	uuid := parsedURL.User.Username()
	if uuid == "" {
		return nil, fmt.Errorf("UUID is required")
	}

	// 解析服务器地址
	host := parsedURL.Hostname()
	if host == "" {
		return nil, fmt.Errorf("server address is required")
	}

	// 解析端口，默认 443
	port := 443
	if portStr := parsedURL.Port(); portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %v", err)
		}
	}

	query := parsedURL.Query()

	// SNI 可由 sni 或 servername 参数指定，优先使用 sni
	sni := query.Get("sni")
	if sni == "" {
		sni = query.Get("servername")
	}

	node := &VlessNode{
		UUID:        uuid,
		Host:        host,
		Port:        port,
		Security:    query.Get("security"),
		Encryption:  query.Get("encryption"),
		Type:        query.Get("type"),
		Flow:        query.Get("flow"),
		PublicKey:   query.Get("pbk"),
		ShortID:     query.Get("sid"),
		SNI:         sni,
		Fingerprint: query.Get("fp"),
		Path:        query.Get("path"),
		HostHeader:  query.Get("host"),
		ServiceName: query.Get("serviceName"),
		Tag:         cleanTag(parsedURL.Fragment),
	}

	// Fragment 解码后为空时回退为默认标签
	if node.Tag == "" {
		node.Tag = "vless-node"
	}

	return node, nil
}

// ConvertVlessToSingBox 将 VlessNode 转换为 sing-box 出站配置。
// 参考：https://sing-box.sagernet.org/zh/configuration/outbound/vless/
func ConvertVlessToSingBox(item *VlessNode) (*entity.SingBoxOut, error) {
	if item == nil {
		return nil, fmt.Errorf("vless node is nil")
	}

	out := &entity.SingBoxOut{
		Type:       "vless",
		Tag:        item.Tag,
		Server:     item.Host,
		ServerPort: item.Port,
		UUID:       item.UUID,
		Flow:       item.Flow,
	}

	// 根据安全层类型配置 TLS
	switch item.Security {
	case "reality":
		// Reality 是基于 TLS 1.3 的伪装安全层，需要服务器公钥和 ShortID
		out.TLS = &entity.SingTLS{
			Enabled:    true,
			ServerName: item.SNI,
			Reality: &entity.SingReality{
				Enabled:   true,
				PublicKey: item.PublicKey,
				ShortID:   item.ShortID,
			},
		}
		// Reality 场景下通常需要 uTLS 指纹伪装以通过 TLS 检测
		if item.Fingerprint != "" {
			out.TLS.Utls = &entity.SingUtls{
				Enabled:     true,
				Fingerprint: item.Fingerprint,
			}
		}
	case "tls":
		// 标准 TLS 模式
		out.TLS = &entity.SingTLS{
			Enabled:    true,
			ServerName: item.SNI,
		}
		// TLS 模式下可选配置 uTLS 指纹伪装
		if item.Fingerprint != "" {
			out.TLS.Utls = &entity.SingUtls{
				Enabled:     true,
				Fingerprint: item.Fingerprint,
			}
		}
	// "none" 或空字符串表示不使用 TLS，不设置 TLS 字段
	}

	// 根据传输协议类型配置 transport
	switch item.Type {
	case "ws":
		// WebSocket 传输，需要配置路径和 Host 头
		transport := &entity.SingTransport{
			Type: "ws",
			Path: item.Path,
		}
		if transport.Path == "" {
			transport.Path = "/"
		}
		if item.HostHeader != "" {
			transport.Headers = map[string][]string{
				"Host": {item.HostHeader},
			}
		}
		out.Transport = transport
	case "grpc":
		// gRPC 传输，需要配置 ServiceName
		transport := &entity.SingTransport{
			Type:        "grpc",
			ServiceName: item.ServiceName,
		}
		out.Transport = transport
	case "httpupgrade":
		// HTTPUpgrade 传输，用于将 HTTP/1.1 连接升级为隧道
		transport := &entity.SingTransport{
			Type: "httpupgrade",
			Path: item.Path,
		}
		if item.HostHeader != "" {
			transport.Host = item.HostHeader
		}
		out.Transport = transport
	// "tcp" 或其他类型不需要额外的传输层配置
	}

	return out, nil
}

// DecodeVlessUrlToSingBox 解析 VLESS URL 并直接返回 sing-box 出站配置。
// 这是供订阅解析链路调用的统一入口。
func DecodeVlessUrlToSingBox(urlStr string) (*entity.SingBoxOut, error) {
	item, err := DecodeVlessUrl(urlStr)
	if err != nil {
		return nil, err
	}
	return ConvertVlessToSingBox(item)
}
