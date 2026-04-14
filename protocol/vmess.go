package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"singboxconfig/entity"
	"strconv"
	"strings"
)

type VmessNode struct {
	Version    string `json:"v"`    // 版本
	Remarks    string `json:"ps"`   // 备注名称
	Address    string `json:"add"`  // 服务器地址
	Port       string `json:"port"` // 端口
	ID         string `json:"id"`   // UUID
	AlterID    string `json:"aid"`  // alter_id
	Network    string `json:"net"`  // 网络类型 (tcp, ws, etc.)
	Type       string `json:"type"` // 伪装类型
	Host       string `json:"host"` // 主机头
	Path       string `json:"path"` // 路径
	TLS        string `json:"tls"`  // 是否启用TLS
	Encryption string `json:"scy"`  // 加密方式，可选字段
}

// 识别vmess的url
// vmess://eyJ2IjoiMiIsInBzIjoiSEsuXHU5OTk5XHU2ZTJmLkMgfCBcdTlhZDhcdTkwMWYuMnhcdTUwMGRcdTczODciLCJhZGQiOiJ4Y3RyYW5zZmVyMXN0LmNva2VjbG91ZC50b3AiLCJwb3J0IjoiMzg2MzAiLCJpZCI6IjNmYzM2NWYyLTk0MGMtNGUwNy05Y2Y3LWExMGM1N2JlODU3YSIsImFpZCI6IjAiLCJuZXQiOiJ0Y3AiLCJ0eXBlIjoibm9uZSIsImhvc3QiOiIiLCJwYXRoIjoiIiwidGxzIjoiIn0=
// {"v":"2","ps":"HK.\u9999\u6e2f.C | \u9ad8\u901f.2x\u500d\u7387","add":"xctransfer1st.cokecloud.top","port":"38630","id":"3fc365f2-940c-4e07-9cf7-a10c57be857a","aid":"0","net":"tcp","type":"none","host":"","path":"","tls":""}
func DecodeVmessUrl(urlStr string) (*VmessNode, error) {
	// 检查 URL 格式
	if !strings.HasPrefix(urlStr, "vmess://") {
		return nil, fmt.Errorf("invalid vmess URL scheme")
	}

	// 提取 base64 编码的部分
	encodedPart := strings.TrimPrefix(urlStr, "vmess://")
	if encodedPart == "" {
		return nil, fmt.Errorf("empty vmess configuration")
	}

	// 解码 base64
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedPart)
	if err != nil {
		// 尝试 URL 安全的 base64 解码
		decodedBytes, err = base64.URLEncoding.DecodeString(encodedPart)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %v", err)
		}
	}

	// 解析 JSON
	var vmessNode VmessNode
	if err := json.Unmarshal(decodedBytes, &vmessNode); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// 验证必需字段
	if vmessNode.Address == "" {
		return nil, fmt.Errorf("server address is required")
	}
	if vmessNode.ID == "" {
		return nil, fmt.Errorf("UUID is required")
	}

	return &vmessNode, nil
}

// 参考https://sing-box.sagernet.org/zh/configuration/outbound/vmess/
func ConvertVmessToSingBox(item *VmessNode) (*entity.SingBoxOut, error) {
	if item == nil {
		return nil, fmt.Errorf("vmess node is nil")
	}

	// 转换端口，提供默认值
	port := 443 // 默认端口
	if item.Port != "" && item.Port != "0" {
		var err error
		port, err = strconv.Atoi(item.Port)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %v", err)
		}
	} else if item.TLS != "tls" {
		port = 80 // 非 TLS 连接默认使用 80 端口
	}

	// 转换 AlterID
	alterID := 0
	if item.AlterID != "" {
		var err error
		alterID, err = strconv.Atoi(item.AlterID)
		if err != nil {
			return nil, fmt.Errorf("invalid alter_id: %v", err)
		}
	}

	// 设置加密方式，默认为 auto
	security := "auto"
	if item.Encryption != "" {
		security = item.Encryption
	}

	// 处理标签
	tag := cleanTag(item.Remarks)
	if tag == "" {
		tag = "vmess-node"
	}

	// 创建基础配置
	out := &entity.SingBoxOut{
		Type:       "vmess",
		Tag:        tag,
		Server:     item.Address,
		ServerPort: port,
		UUID:       item.ID,
		AlterID:    alterID,
		Security:   security,
		Multiplex:  &entity.SingMultiplex{},
	}

	// 处理 TLS 配置
	if item.TLS == "tls" {
		out.TLS = &entity.SingTLS{
			Enabled: true,
		}
		if item.Host != "" {
			out.TLS.ServerName = item.Host
		}
	}

	// 处理传输协议
	transport := &entity.SingTransport{}
	switch item.Network {
	case "ws":
		transport.Type = "ws"
		if item.Path != "" {
			transport.Path = item.Path
		} else {
			transport.Path = "/" // 默认路径
		}
		if item.Host != "" {
			transport.Headers = map[string][]string{
				"Host": {item.Host},
			}
		}
	case "h2", "http":
		transport.Type = "http"
		if item.Path != "" {
			transport.Path = item.Path
		} else {
			transport.Path = "/" // 默认路径
		}
		if item.Host != "" {
			// 支持多个主机，以逗号分隔
			hosts := strings.Split(item.Host, ",")
			for i := range hosts {
				hosts[i] = strings.TrimSpace(hosts[i])
			}
			transport.Host = hosts
		}
	case "grpc":
		transport.Type = "grpc"
		if item.Path != "" {
			transport.ServiceName = item.Path
		} else {
			transport.ServiceName = "GunService" // 默认服务名
		}
	case "quic":
		// QUIC 传输通常不需要额外的传输配置
		// 但可能需要特殊的 TLS 配置
		transport = nil
	case "kcp":
		// KCP 协议支持
		transport.Type = "kcp"
		// KCP 通常有自己的配置参数，但 sing-box 可能不完全支持
	case "tcp":
		// TCP 不需要特殊配置，但如果有伪装类型需要处理
		if item.Type == "http" {
			transport.Type = "http"
			if item.Path != "" {
				transport.Path = item.Path
			} else {
				transport.Path = "/" // 默认路径
			}
			if item.Host != "" {
				transport.Headers = map[string][]string{
					"Host": {item.Host},
				}
			}
		}
	default:
		// 对于未知的网络类型，不设置传输配置
		transport = nil
	}

	// 只有在有实际配置时才设置 transport
	if transport != nil && (transport.Type != "" || transport.Path != "" || len(transport.Headers) > 0) {
		out.Transport = transport
	}

	return out, nil
}

func DecodeVmessUrlToSingBox(urlStr string) (*entity.SingBoxOut, error) {
	item, err := DecodeVmessUrl(urlStr)
	if err != nil {
		return nil, err
	}
	return ConvertVmessToSingBox(item)
}
