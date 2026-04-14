package protocol

import (
	"fmt"
	"net/url"
	"singboxconfig/entity"
	"strconv"
	"strings"
	"unicode"
)

type TrojanNode struct {
	Scheme        string `json:"scheme"`         // 协议名称，固定为 "trojan"
	Password      string `json:"password"`       // 用户认证密码（UUID 格式）
	Host          string `json:"host"`           // 服务器地址（域名或 IP）
	Port          int    `json:"port"`           // 服务器端口号
	Security      string `json:"security"`       // 加密方式，一般为 "tls"
	SNI           string `json:"sni"`            // TLS 握手时使用的 SNI（Server Name Indication）
	AllowInsecure bool   `json:"allow_insecure"` // 是否允许不安全的 TLS 连接（跳过证书验证）
	Peer          string `json:"peer"`           // 指定的 TLS 服务器名称，用于验证证书
	Type          string `json:"type"`           // 传输类型，一般为 "tcp"
	HostHeader    string `json:"host_header"`    // 伪装主机头（用于 WebSocket 或 TLS 中的 Host 字段）
	Tag           string `json:"tag"`            // 节点备注（用户可读标签，例如地区、用途）
}

func DecodeTrojanUrl(urlStr string) (*TrojanNode, error) {
	// 解析 URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	// 验证协议
	if parsedURL.Scheme != "trojan" {
		return nil, fmt.Errorf("invalid scheme: %s", parsedURL.Scheme)
	}

	// 获取密码（用户名部分）
	password := parsedURL.User.Username()
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// 获取主机和端口
	host := parsedURL.Hostname()
	portStr := parsedURL.Port()
	port := 443 // 默认端口
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %v", err)
		}
	}

	// 解析查询参数
	query := parsedURL.Query()

	// 创建 TrojanNode
	item := &TrojanNode{
		Scheme:        "trojan",
		Password:      password,
		Host:          host,
		Port:          port,
		Security:      query.Get("security"),
		SNI:           query.Get("sni"),
		AllowInsecure: query.Get("allowInsecure") == "1",
		Peer:          query.Get("peer"),
		Type:          query.Get("type"),
		HostHeader:    query.Get("host"),
		Tag:           cleanTag(parsedURL.Fragment),
	}

	return item, nil
}

// 参考：https://sing-box.sagernet.org/zh/configuration/outbound/trojan/
func ConvertTrojanToSingBox(item *TrojanNode) (*entity.SingBoxOut, error) {
	if item == nil {
		return nil, fmt.Errorf("trojan node is nil")
	}

	// 创建 TLS 配置
	tlsConfig := &entity.SingTLS{
		ServerName: item.SNI,
	}

	// 默认启用不安全连接
	tlsConfig.Insecure = true

	if item.Security != "" && item.Security != "tls" {
		tlsConfig.Enabled = false
	} else {
		tlsConfig.Enabled = true
	}

	// 如果设置了 peer，则使用 peer 作为 ServerName
	if item.Peer != "" {
		tlsConfig.ServerName = item.Peer
	}

	if item.Type == "" {
		item.Type = "tcp"
	}

	// 创建 sing-box 出站配置
	out := &entity.SingBoxOut{
		Type:       "trojan",
		Tag:        item.Tag,
		Server:     item.Host,
		ServerPort: item.Port,
		Password:   item.Password,
		Network:    item.Type,
		TLS:        tlsConfig,
		Multiplex:  &entity.SingMultiplex{},
		Transport:  &entity.SingTransport{},
	}

	return out, nil
}

func DecodeTrojanUrlToSingBox(urlStr string) (*entity.SingBoxOut, error) {
	item, err := DecodeTrojanUrl(urlStr)
	if err != nil {
		return nil, err
	}
	return ConvertTrojanToSingBox(item)
}

// cleanTag 清理 tag 中的特殊字符、emoji、空格和换行符
func cleanTag(tag string) string {
	// 移除换行符
	tag = strings.ReplaceAll(tag, "\n", "")
	tag = strings.ReplaceAll(tag, "\r", "")

	// 移除空格
	tag = strings.TrimSpace(tag)

	// 移除特殊字符和 emoji
	var result strings.Builder
	for _, r := range tag {
		if unicode.IsPrint(r) && !unicode.IsSpace(r) && !unicode.IsSymbol(r) && !unicode.IsPunct(r) {
			result.WriteRune(r)
		}
	}

	return strings.TrimPrefix(result.String(), "")
}
