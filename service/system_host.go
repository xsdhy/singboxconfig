package service

import (
	"errors"
	"net/url"
	"singboxconfig/storage"
	"strings"
)

// systemHostSettingKey 是“系统 Host”在全局设置存储中的固定 key。
// 它与 dnsConfigSettingKey 同级，记录本服务对外可访问的基础地址（如 https://config.example.com），
// 用于把规则集 open 接口拼接成绝对 URL，供生成链路引用。
const systemHostSettingKey = "system_host"

// errInvalidSystemHost 表示提交的系统 Host 不是合法的 http/https 绝对地址。
var errInvalidSystemHost = errors.New("system_host must be an absolute http(s) url, e.g. https://config.example.com")

// normalizeSystemHost 统一清洗系统 Host：去掉首尾空白与尾部所有斜杠。
// 读取与保存都走这里，保证存储值与拼接 URL 时的基地址一致。
func normalizeSystemHost(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

// validateSystemHost 校验系统 Host 是否为合法的 http/https 绝对 URL。
// 允许空串（表示未配置，生成时回退到展开/inline 行为）；非空时必须能解析、scheme 为 http/https 且带主机名。
// 该函数是纯函数，便于直接单元测试各种合法/非法输入。
func validateSystemHost(raw string) error {
	host := normalizeSystemHost(raw)
	if host == "" {
		return nil
	}
	parsed, err := url.Parse(host)
	if err != nil {
		return errInvalidSystemHost
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return errInvalidSystemHost
	}
	if parsed.Host == "" {
		return errInvalidSystemHost
	}
	return nil
}

// resolveSystemHost 读取后台保存的系统 Host，并做规范化清洗。// 未配置时返回空串，让生成链路回退到展开/inline 行为；存储中保存了非法值时同样按未配置处理，
// 避免生成指向坏地址的远程规则集 URL（保存接口已拒绝非法值，这里是额外的防御性兜底）。
func (s *Service) resolveSystemHost() (string, error) {
	value, err := s.storage.GetGlobalSetting(systemHostSettingKey)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	host := normalizeSystemHost(value)
	if validateSystemHost(host) != nil {
		return "", nil
	}
	return host, nil
}

// normalizeGlobalSettingValue 在写入全局设置前，按 key 做针对性的规范化与校验。
// 目前仅 system_host 需要特殊处理：去尾斜杠并校验为合法 http/https 绝对地址，非法值返回错误以便拒绝保存。
// 其余 key 原样返回，保持通用 key/value 设置的写入行为不变。
func normalizeGlobalSettingValue(key string, value string) (string, error) {
	if key == systemHostSettingKey {
		host := normalizeSystemHost(value)
		if err := validateSystemHost(host); err != nil {
			return "", err
		}
		return host, nil
	}
	return value, nil
}
