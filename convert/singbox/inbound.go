package singbox

import (
	"encoding/json"
	"singboxconfig/entity"
	"sort"

	"github.com/sirupsen/logrus"
)

// GetInbounds 将存储中的 Inbound 模板转换成 sing-box 入站配置。
// 会按 Sort 排序，跳过停用项，并忽略无法反序列化的脏数据。
func GetInbounds(items []*entity.Inbound) []entity.SingInbound {
	if len(items) == 0 {
		return nil
	}

	sorted := append([]*entity.Inbound(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Sort == sorted[j].Sort {
			return sorted[i].Tag < sorted[j].Tag
		}
		return sorted[i].Sort < sorted[j].Sort
	})

	inbounds := make([]entity.SingInbound, 0, len(sorted))
	for _, item := range sorted {
		if item == nil || !item.Enabled {
			continue
		}

		var inbound entity.SingInbound
		if err := json.Unmarshal([]byte(item.ConfigJSON), &inbound); err != nil {
			logrus.WithError(err).WithField("tag", item.Tag).Warn("GetInbounds: invalid inbound config, skip")
			continue
		}
		inbounds = append(inbounds, inbound)
	}

	return inbounds
}
