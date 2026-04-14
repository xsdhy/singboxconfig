package protocol

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"singboxconfig/entity"
	"strconv"
	"strings"
)

type SSNode struct {
	Method   string
	Password string
	Host     string
	Port     int
	Group    string
	Tag      string
}

func DecodeSSURL(ssURL string) (*SSNode, error) {
	if !strings.HasPrefix(ssURL, "ss://") {
		return nil, fmt.Errorf("invalid ss:// URL")
	}
	ssURL = ssURL[len("ss://"):]

	// 解析 Base64 编码部分和主机端口部分
	parts := strings.SplitN(ssURL, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ss:// URL")
	}

	// 尝试多种 base64 编码方式
	var decoded []byte
	var decodeErr error

	// 先尝试标准编码（带 padding）
	decoded, decodeErr = base64.StdEncoding.DecodeString(parts[0])
	if decodeErr != nil {
		// 再尝试 URL 编码
		decoded, decodeErr = base64.URLEncoding.DecodeString(parts[0])
		if decodeErr != nil {
			// 最后尝试 RawURL 编码
			decoded, decodeErr = base64.RawURLEncoding.DecodeString(parts[0])
			if decodeErr != nil {
				return nil, fmt.Errorf("Base64 decode error: %v", decodeErr)
			}
		}
	}

	// 解析方法和密码
	methodPassword := strings.SplitN(string(decoded), ":", 2)
	if len(methodPassword) != 2 {
		return nil, fmt.Errorf("invalid method:password format")
	}
	method := methodPassword[0]
	password := methodPassword[1]

	// 解析主机端口部分和其他参数
	// 支持两种格式: host:port/?params#tag 和 host:port?params#tag
	hostPortAndParams := parts[1]

	// 先尝试找到查询参数或片段的起始位置
	queryStart := strings.Index(hostPortAndParams, "?")
	fragmentStart := strings.Index(hostPortAndParams, "#")
	slashStart := strings.Index(hostPortAndParams, "/")

	var hostPort string
	group := ""
	tag := ""

	// 确定 hostPort 的结束位置
	endPos := len(hostPortAndParams)
	if queryStart != -1 && queryStart < endPos {
		endPos = queryStart
	}
	if fragmentStart != -1 && fragmentStart < endPos {
		endPos = fragmentStart
	}
	if slashStart != -1 && slashStart < endPos {
		endPos = slashStart
	}

	hostPort = hostPortAndParams[:endPos]

	// 解析剩余部分的查询参数和片段
	if endPos < len(hostPortAndParams) {
		remaining := hostPortAndParams[endPos:]
		// 如果以 / 开头，去掉它
		if strings.HasPrefix(remaining, "/") {
			remaining = remaining[1:]
		}
		if remaining != "" {
			// 构造一个完整的 URL 用于解析
			var parseURL string
			if strings.HasPrefix(remaining, "?") {
				parseURL = "http://dummy/" + remaining
			} else {
				parseURL = "http://dummy/?" + remaining
			}
			u, parseErr := url.Parse(parseURL)
			if parseErr == nil {
				group = u.Query().Get("group")
				tag = u.Fragment
			}
		}
	}

	host, portStr, splitErr := net.SplitHostPort(hostPort)
	if splitErr != nil {
		return nil, fmt.Errorf("host:port split error: %v", splitErr)
	}
	port, parseErr := strconv.Atoi(portStr)
	if parseErr != nil {
		return nil, fmt.Errorf("port parse error: %v", parseErr)
	}

	return &SSNode{
		Method:   method,
		Password: password,
		Host:     host,
		Port:     port,
		Group:    group,
		Tag:      tag,
	}, nil
}

// 参考：https://sing-box.sagernet.org/zh/configuration/outbound/shadowsocks/
func ConvertSSToSingBox(item *SSNode) (*entity.SingBoxOut, error) {
	if item == nil {
		return nil, fmt.Errorf("ss node is nil")
	}

	return &entity.SingBoxOut{
		Type:       "shadowsocks",
		Tag:        cleanTag(item.Tag),
		Server:     item.Host,
		ServerPort: item.Port,
		Method:     item.Method,
		Password:   item.Password,
		Network:    "tcp",                   // 如需支持 udp，可根据实际情况调整
		Plugin:     "",                      // 可根据需要补充
		PluginOpts: "",                      // 可根据需要补充
		UdpOverTcp: nil,                     // 可根据需要补充
		Multiplex:  &entity.SingMultiplex{}, // 可根据需要补充
	}, nil
}

func DecodeSSURLToSingBox(urlStr string) (*entity.SingBoxOut, error) {
	item, err := DecodeSSURL(urlStr)
	if err != nil {
		return nil, err
	}
	return ConvertSSToSingBox(item)
}
