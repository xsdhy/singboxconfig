package singbox

import (
	"encoding/json"
	"singboxconfig/entity"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetExtraOutbounds 将存储里的额外 Outbound 模板转换成可直接注入生成链路的出站配置。
func GetExtraOutbounds(deviceCode string, items []*entity.Outbound) []entity.SingBoxOut {
	if len(items) == 0 {
		return nil
	}

	sorted := append([]*entity.Outbound(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Sort == sorted[j].Sort {
			return sorted[i].Tag < sorted[j].Tag
		}
		return sorted[i].Sort < sorted[j].Sort
	})

	outbounds := make([]entity.SingBoxOut, 0, len(sorted))
	for _, item := range sorted {
		if item == nil || !item.Enabled || !isExtraOutboundVisibleForDevice(item.VisibleDevices, deviceCode) {
			continue
		}

		var outbound entity.SingBoxOut
		if err := json.Unmarshal([]byte(item.ConfigJSON), &outbound); err != nil {
			logrus.WithError(err).WithField("tag", item.Tag).Warn("GetExtraOutbounds: invalid extra outbound config, skip")
			continue
		}
		outbounds = append(outbounds, outbound)
	}

	return outbounds
}

func isExtraOutboundVisibleForDevice(visibleDevices string, deviceCode string) bool {
	if strings.TrimSpace(visibleDevices) == "" {
		return true
	}

	for _, item := range strings.Split(visibleDevices, ",") {
		if strings.TrimSpace(item) == deviceCode {
			return true
		}
	}
	return false
}
