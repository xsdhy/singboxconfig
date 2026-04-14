package protocol

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"singboxconfig/entity"
	"strconv"
	"strings"
)

type SSRNode struct {
	Server        string
	Port          int
	Protocol      string
	Method        string
	Obfs          string
	Password      string
	ObfsParam     string
	ProtocolParam string
	Group         string
	Remarks       string
	Tag           string
}

// DecodeSSRURL 解析 SSR URL
// 支持两种格式：
// 1. ssr://base64(server:port:protocol:method:obfs:password/?query_params)
// 2. ssr://server:port:protocol:method:obfs:password/?query_params (直接格式)
func DecodeSSRURL(ssrURL string) (*SSRNode, error) {
	if !strings.HasPrefix(ssrURL, "ssr://") {
		return nil, fmt.Errorf("invalid ssr:// URL")
	}
	ssrURL = ssrURL[len("ssr://"):]

	// 首先尝试作为base64编码的格式解析
	content, err := decodeBase64WithPadding(ssrURL)
	if err == nil && content != ssrURL {
		// 成功解码且内容发生了变化，说明是base64编码的
		return parseSSRContent(content)
	}

	// 如果base64解码失败或内容没有变化，尝试作为直接格式解析
	return parseSSRContent(ssrURL)
}

func parseSSRContent(content string) (*SSRNode, error) {
	// URL解码内容
	decodedContent, err := url.QueryUnescape(content)
	if err == nil {
		content = decodedContent
	}

	// 分离主要部分和查询参数
	parts := strings.SplitN(content, "/?", 2)
	mainPart := parts[0]

	// 解析主要部分: server:port:protocol:method:obfs:password
	mainFields := strings.Split(mainPart, ":")
	if len(mainFields) != 6 {
		return nil, fmt.Errorf("invalid SSR format, expected 6 fields, got %d", len(mainFields))
	}

	server := mainFields[0]
	portStr := mainFields[1]
	protocol := mainFields[2]
	method := mainFields[3]
	obfs := mainFields[4]
	passwordB64 := mainFields[5]

	// 解析端口
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	// 解码密码
	password, _ := decodeBase64WithPadding(passwordB64)

	node := &SSRNode{
		Server:   server,
		Port:     port,
		Protocol: protocol,
		Method:   method,
		Obfs:     obfs,
		Password: password,
	}

	// 解析查询参数
	if len(parts) == 2 && parts[1] != "" {
		if err := parseSSRQueryParams(node, parts[1]); err != nil {
			return nil, fmt.Errorf("failed to parse query params: %v", err)
		}
	}

	return node, nil
}

// decodeBase64WithPadding 解码base64字符串，自动添加padding
func decodeBase64WithPadding(encoded string) (string, error) {
	// 检查是否像base64字符串（只包含base64字符）
	if !isBase64Like(encoded) {
		return encoded, fmt.Errorf("not base64 like")
	}

	// 尝试URL安全的base64解码
	if decoded, err := base64.URLEncoding.DecodeString(encoded); err == nil {
		return string(decoded), nil
	}

	// 添加padding后再尝试
	paddedEncoded := addBase64Padding(encoded)
	if decoded, err := base64.URLEncoding.DecodeString(paddedEncoded); err == nil {
		return string(decoded), nil
	}

	// 尝试标准base64解码
	if decoded, err := base64.StdEncoding.DecodeString(paddedEncoded); err == nil {
		return string(decoded), nil
	}

	// 如果都失败了，返回原字符串
	return encoded, fmt.Errorf("base64 decode failed")
}

// isBase64Like 检查字符串是否像base64编码
func isBase64Like(s string) bool {
	// base64字符只包含 A-Za-z0-9+/_-=
	for _, r := range s {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') || r == '+' || r == '/' || r == '_' || r == '-' || r == '=') {
			return false
		}
	}
	// 长度至少应该是4的倍数（或接近）才像base64
	return len(s) >= 4
}

// addBase64Padding 为base64字符串添加必要的padding
func addBase64Padding(encoded string) string {
	if len(encoded)%4 == 0 {
		return encoded
	}
	return encoded + strings.Repeat("=", 4-len(encoded)%4)
}

func parseSSRQueryParams(node *SSRNode, queryStr string) error {
	params, err := url.ParseQuery(queryStr)
	if err != nil {
		return err
	}

	// 解析 obfsparam
	if obfsParam := params.Get("obfsparam"); obfsParam != "" {
		decoded, _ := decodeBase64WithPadding(obfsParam)
		node.ObfsParam = decoded
	}

	// 解析 protoparam
	if protoParam := params.Get("protoparam"); protoParam != "" {
		decoded, _ := decodeBase64WithPadding(protoParam)
		node.ProtocolParam = decoded
	}

	// 解析 remarks
	if remarks := params.Get("remarks"); remarks != "" {
		decoded, _ := decodeBase64WithPadding(remarks)
		node.Remarks = decoded
		node.Tag = node.Remarks
	}

	// 解析 group
	if group := params.Get("group"); group != "" {
		decoded, _ := decodeBase64WithPadding(group)
		node.Group = decoded
	}

	return nil
}

// ConvertSSRToSingBox 将 SSR 节点转换为 SingBox 配置
// 参考：https://sing-box.sagernet.org/zh/configuration/outbound/shadowsocksr/
func ConvertSSRToSingBox(item *SSRNode) (*entity.SingBoxOut, error) {
	if item == nil {
		return nil, fmt.Errorf("ssr node is nil")
	}

	// 创建 SingObfs 结构体
	var obfs *entity.SingObfs
	if item.Obfs != "" {
		obfs = &entity.SingObfs{
			Value: item.Obfs,
			Type:  "", // SSR的obfs通常不需要额外的type
		}
	}

	return &entity.SingBoxOut{
		Type:          "shadowsocksr",
		Tag:           cleanTag(item.Tag),
		Server:        item.Server,
		ServerPort:    item.Port,
		Method:        item.Method,
		Password:      item.Password,
		Protocol:      item.Protocol,
		ProtocolParam: item.ProtocolParam,
		Obfs:          obfs,
		ObfsParam:     item.ObfsParam,
		Network:       "tcp",                   // 默认使用 tcp
		Multiplex:     &entity.SingMultiplex{}, // 可根据需要调整
	}, nil
}

// DecodeSSRURLToSingBox 直接将 SSR URL 转换为 SingBox 配置
func DecodeSSRURLToSingBox(urlStr string) (*entity.SingBoxOut, error) {
	item, err := DecodeSSRURL(urlStr)
	if err != nil {
		return nil, err
	}
	return ConvertSSRToSingBox(item)
}
