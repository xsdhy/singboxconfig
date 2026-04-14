package storage

import (
	"errors"
	"singboxconfig/entity"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DatabaseStorage 数据库存储实现
type DatabaseStorage struct {
	db *gorm.DB
}

// NewDatabaseStorage 创建数据库存储实例
func NewDatabaseStorage(db *gorm.DB) (*DatabaseStorage, error) {
	// 需求明确说明当前无需保留历史 ExtraOutbound 表，启动时直接清理旧表结构。
	if db.Migrator().HasTable("extra_outbounds") {
		if err := db.Migrator().DropTable("extra_outbounds"); err != nil {
			return nil, err
		}
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&DBSubscribe{},
		&DBNodeGroup{},
		&DBRuleSet{},
		&DBGlobalSetting{},
		&DBDevice{},
		&DBInbound{},
		&DBDeviceInbound{},
		&DBWireGuard{},
		&DBWireGuardPeer{},
		&DBOutbound{},
	); err != nil {
		return nil, err
	}

	return &DatabaseStorage{db: db}, nil
}

// Subscribe 相关方法
func (ds *DatabaseStorage) CreateSubscribe(subscribe *entity.Subscribe) error {
	dbSubscribe := &DBSubscribe{}
	dbSubscribe.FromEntity(subscribe)
	return ds.db.Create(dbSubscribe).Error
}

func (ds *DatabaseStorage) GetSubscribe(name string) (*entity.Subscribe, error) {
	var dbSubscribe DBSubscribe
	err := ds.db.Where("name = ?", name).First(&dbSubscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return dbSubscribe.ToEntity(), nil
}

func (ds *DatabaseStorage) ListSubscribes() ([]*entity.Subscribe, error) {
	var dbSubscribes []DBSubscribe
	if err := ds.db.Find(&dbSubscribes).Error; err != nil {
		return nil, err
	}

	subscribes := make([]*entity.Subscribe, 0, len(dbSubscribes))
	for _, db := range dbSubscribes {
		subscribes = append(subscribes, db.ToEntity())
	}
	return subscribes, nil
}

func (ds *DatabaseStorage) UpdateSubscribe(subscribe *entity.Subscribe) error {
	dbSubscribe := &DBSubscribe{}
	dbSubscribe.FromEntity(subscribe)
	result := ds.db.Model(&DBSubscribe{}).Where("name = ?", subscribe.Name).Updates(dbSubscribe)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ds *DatabaseStorage) DeleteSubscribe(name string) error {
	result := ds.db.Where("name = ?", name).Delete(&DBSubscribe{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// NodeGroup 相关方法
func (ds *DatabaseStorage) CreateNodeGroup(group *entity.NodeGroup) error {
	dbGroup := &DBNodeGroup{}
	dbGroup.FromEntity(group)
	return ds.db.Create(dbGroup).Error
}

func (ds *DatabaseStorage) GetNodeGroup(tag string) (*entity.NodeGroup, error) {
	var dbGroup DBNodeGroup
	err := ds.db.Where("tag = ?", tag).First(&dbGroup).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return dbGroup.ToEntity(), nil
}

func (ds *DatabaseStorage) ListNodeGroups() ([]*entity.NodeGroup, error) {
	var dbGroups []DBNodeGroup
	if err := ds.db.Find(&dbGroups).Error; err != nil {
		return nil, err
	}

	groups := make([]*entity.NodeGroup, 0, len(dbGroups))
	for _, db := range dbGroups {
		groups = append(groups, db.ToEntity())
	}
	return groups, nil
}

func (ds *DatabaseStorage) UpdateNodeGroup(group *entity.NodeGroup) error {
	dbGroup := &DBNodeGroup{}
	dbGroup.FromEntity(group)
	result := ds.db.Model(&DBNodeGroup{}).Where("tag = ?", group.Tag).Updates(dbGroup)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ds *DatabaseStorage) DeleteNodeGroup(tag string) error {
	result := ds.db.Where("tag = ?", tag).Delete(&DBNodeGroup{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// RuleSet 相关方法
func (ds *DatabaseStorage) CreateRuleSet(ruleSet *entity.RuleSet) error {
	dbRuleSet := &DBRuleSet{}
	dbRuleSet.FromEntity(ruleSet)
	return ds.db.Create(dbRuleSet).Error
}

func (ds *DatabaseStorage) GetRuleSet(tag string) (*entity.RuleSet, error) {
	var dbRuleSet DBRuleSet
	err := ds.db.Where("tag = ?", tag).First(&dbRuleSet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return dbRuleSet.ToEntity(), nil
}

func (ds *DatabaseStorage) ListRuleSets() ([]*entity.RuleSet, error) {
	var dbRuleSets []DBRuleSet
	if err := ds.db.Find(&dbRuleSets).Error; err != nil {
		return nil, err
	}

	ruleSets := make([]*entity.RuleSet, 0, len(dbRuleSets))
	for _, db := range dbRuleSets {
		ruleSets = append(ruleSets, db.ToEntity())
	}
	return ruleSets, nil
}

func (ds *DatabaseStorage) UpdateRuleSet(ruleSet *entity.RuleSet) error {
	dbRuleSet := &DBRuleSet{}
	dbRuleSet.FromEntity(ruleSet)
	result := ds.db.Model(&DBRuleSet{}).Where("tag = ?", ruleSet.Tag).Updates(dbRuleSet)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ds *DatabaseStorage) DeleteRuleSet(tag string) error {
	result := ds.db.Where("tag = ?", tag).Delete(&DBRuleSet{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// GlobalSetting 相关方法
func (ds *DatabaseStorage) SetGlobalSetting(key, value string) error {
	dbSetting := &DBGlobalSetting{
		Key:   key,
		Value: value,
	}
	// 使用 UPSERT 语义：存在则更新，不存在则创建
	return ds.db.Save(dbSetting).Error
}

func (ds *DatabaseStorage) GetGlobalSetting(key string) (string, error) {
	var dbSetting DBGlobalSetting
	err := ds.db.Where("key = ?", key).First(&dbSetting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return dbSetting.Value, nil
}

func (ds *DatabaseStorage) ListGlobalSettings() (map[string]string, error) {
	var dbSettings []DBGlobalSetting
	if err := ds.db.Find(&dbSettings).Error; err != nil {
		return nil, err
	}

	settings := make(map[string]string)
	for _, db := range dbSettings {
		settings[db.Key] = db.Value
	}
	return settings, nil
}

func (ds *DatabaseStorage) DeleteGlobalSetting(key string) error {
	result := ds.db.Where("key = ?", key).Delete(&DBGlobalSetting{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Device 相关方法
func (ds *DatabaseStorage) CreateDevice(device *entity.Device) error {
	dbDevice := &DBDevice{}
	dbDevice.FromEntity(device)
	return ds.db.Create(dbDevice).Error
}

func (ds *DatabaseStorage) GetDevice(code string) (*entity.Device, error) {
	var dbDevice DBDevice
	err := ds.db.Where("code = ?", code).First(&dbDevice).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return dbDevice.ToEntity(), nil
}

func (ds *DatabaseStorage) ListDevices() ([]*entity.Device, error) {
	var dbDevices []DBDevice
	if err := ds.db.Find(&dbDevices).Error; err != nil {
		return nil, err
	}

	devices := make([]*entity.Device, 0, len(dbDevices))
	for _, db := range dbDevices {
		devices = append(devices, db.ToEntity())
	}
	return devices, nil
}

func (ds *DatabaseStorage) UpdateDevice(device *entity.Device) error {
	dbDevice := &DBDevice{}
	dbDevice.FromEntity(device)
	result := ds.db.Model(&DBDevice{}).Where("code = ?", device.Code).Updates(dbDevice)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ds *DatabaseStorage) DeleteDevice(code string) error {
	return ds.db.Transaction(func(tx *gorm.DB) error {
		// 先清理绑定关系，再删除设备主体，兼容未配置级联删除的数据库。
		if err := tx.Where("device_code = ?", code).Delete(&DBDeviceInbound{}).Error; err != nil {
			return err
		}
		result := tx.Where("code = ?", code).Delete(&DBDevice{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// Inbound 相关方法
func (ds *DatabaseStorage) CreateInbound(inbound *entity.Inbound) error {
	dbInbound := &DBInbound{}
	dbInbound.FromEntity(inbound)
	return ds.db.Create(dbInbound).Error
}

func (ds *DatabaseStorage) GetInbound(tag string) (*entity.Inbound, error) {
	var dbInbound DBInbound
	err := ds.db.Where("tag = ?", tag).First(&dbInbound).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return dbInbound.ToEntity(), nil
}

func (ds *DatabaseStorage) ListInbounds() ([]*entity.Inbound, error) {
	var dbInbounds []DBInbound
	if err := ds.db.Find(&dbInbounds).Error; err != nil {
		return nil, err
	}

	inbounds := make([]*entity.Inbound, 0, len(dbInbounds))
	for _, db := range dbInbounds {
		inbounds = append(inbounds, db.ToEntity())
	}
	return inbounds, nil
}

func (ds *DatabaseStorage) UpdateInbound(inbound *entity.Inbound) error {
	dbInbound := &DBInbound{}
	dbInbound.FromEntity(inbound)
	result := ds.db.Model(&DBInbound{}).Where("tag = ?", inbound.Tag).Updates(dbInbound)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ds *DatabaseStorage) DeleteInbound(tag string) error {
	return ds.db.Transaction(func(tx *gorm.DB) error {
		// 先清理设备绑定，避免外键限制阻塞 Inbound 删除。
		if err := tx.Where("inbound_tag = ?", tag).Delete(&DBDeviceInbound{}).Error; err != nil {
			return err
		}
		result := tx.Where("tag = ?", tag).Delete(&DBInbound{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// DeviceInbound 相关方法
func (ds *DatabaseStorage) SetDeviceInbounds(deviceCode string, bindings []*entity.DeviceInbound) error {
	return ds.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("device_code = ?", deviceCode).Delete(&DBDeviceInbound{}).Error; err != nil {
			return err
		}
		for _, binding := range bindings {
			if binding == nil {
				continue
			}
			dbBinding := &DBDeviceInbound{}
			dbBinding.FromEntity(binding)
			if err := tx.Create(dbBinding).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (ds *DatabaseStorage) ListDeviceInbounds(deviceCode string) ([]*entity.DeviceInbound, error) {
	var dbBindings []DBDeviceInbound
	if err := ds.db.Where("device_code = ?", deviceCode).Order("sort asc").Find(&dbBindings).Error; err != nil {
		return nil, err
	}

	bindings := make([]*entity.DeviceInbound, 0, len(dbBindings))
	for _, db := range dbBindings {
		bindings = append(bindings, db.ToEntity())
	}
	return bindings, nil
}

// WireGuard 相关方法
func (ds *DatabaseStorage) CreateWireGuard(item *entity.WireGuard) error {
	dbItem := &DBWireGuard{}
	dbItem.FromEntity(item)
	return ds.db.Create(dbItem).Error
}

func (ds *DatabaseStorage) GetWireGuard(tag string) (*entity.WireGuard, error) {
	var dbItem DBWireGuard
	err := ds.db.Where("tag = ?", tag).First(&dbItem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return dbItem.ToEntity(), nil
}

func (ds *DatabaseStorage) ListWireGuards() ([]*entity.WireGuard, error) {
	var dbItems []DBWireGuard
	if err := ds.db.Find(&dbItems).Error; err != nil {
		return nil, err
	}

	items := make([]*entity.WireGuard, 0, len(dbItems))
	for _, db := range dbItems {
		items = append(items, db.ToEntity())
	}
	return items, nil
}

func (ds *DatabaseStorage) UpdateWireGuard(item *entity.WireGuard) error {
	dbItem := &DBWireGuard{}
	dbItem.FromEntity(item)
	result := ds.db.Model(&DBWireGuard{}).Where("tag = ?", item.Tag).Updates(dbItem)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ds *DatabaseStorage) DeleteWireGuard(tag string) error {
	return ds.db.Transaction(func(tx *gorm.DB) error {
		// 删除模板前先清理 peer 和设备引用，避免留下悬挂的 WireGuardTag。
		if err := tx.Model(&DBDevice{}).
			Where("wire_guard_tag = ?", tag).
			Updates(map[string]interface{}{
				"wire_guard_tag":         "",
				"wire_guard_client_addr": "",
				"wire_guard_client_key":  "",
			}).Error; err != nil {
			return err
		}
		if err := tx.Where("wire_guard_tag = ?", tag).Delete(&DBWireGuardPeer{}).Error; err != nil {
			return err
		}
		result := tx.Where("tag = ?", tag).Delete(&DBWireGuard{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// WireGuardPeer 相关方法
func (ds *DatabaseStorage) CreateWireGuardPeer(item *entity.WireGuardPeer) error {
	dbItem := &DBWireGuardPeer{}
	dbItem.FromEntity(item)
	if err := ds.db.Create(dbItem).Error; err != nil {
		return err
	}
	item.ID = dbItem.ID
	return nil
}

func (ds *DatabaseStorage) ListWireGuardPeers(wireGuardTag string) ([]*entity.WireGuardPeer, error) {
	var dbItems []DBWireGuardPeer
	if err := ds.db.Where("wire_guard_tag = ?", wireGuardTag).Order("sort asc, id asc").Find(&dbItems).Error; err != nil {
		return nil, err
	}

	items := make([]*entity.WireGuardPeer, 0, len(dbItems))
	for _, db := range dbItems {
		items = append(items, db.ToEntity())
	}
	return items, nil
}

func (ds *DatabaseStorage) UpdateWireGuardPeer(item *entity.WireGuardPeer) error {
	dbItem := &DBWireGuardPeer{}
	dbItem.FromEntity(item)
	result := ds.db.Model(&DBWireGuardPeer{}).Where("id = ?", item.ID).Updates(dbItem)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ds *DatabaseStorage) DeleteWireGuardPeer(id int64) error {
	result := ds.db.Where("id = ?", id).Delete(&DBWireGuardPeer{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ExtraOutbound 相关方法
func (ds *DatabaseStorage) CreateExtraOutbound(outbound *entity.Outbound) error {
	outbound.Source = entity.OutboundSourceManual
	outbound.SubscribeName = ""
	return ds.db.Model(&DBOutbound{}).Create(databaseOutboundValues(outbound)).Error
}

func (ds *DatabaseStorage) GetExtraOutbound(tag string) (*entity.Outbound, error) {
	var dbOutbound DBOutbound
	err := ds.db.Where("tag = ? AND source = ?", tag, entity.OutboundSourceManual).First(&dbOutbound).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return dbOutbound.ToEntity(), nil
}

func (ds *DatabaseStorage) ListExtraOutbounds() ([]*entity.Outbound, error) {
	var dbOutbounds []DBOutbound
	if err := ds.db.Where("source = ?", entity.OutboundSourceManual).Find(&dbOutbounds).Error; err != nil {
		return nil, err
	}

	outbounds := make([]*entity.Outbound, 0, len(dbOutbounds))
	for _, db := range dbOutbounds {
		outbounds = append(outbounds, db.ToEntity())
	}
	return outbounds, nil
}

func (ds *DatabaseStorage) UpdateExtraOutbound(outbound *entity.Outbound) error {
	outbound.Source = entity.OutboundSourceManual
	outbound.SubscribeName = ""
	result := ds.db.Model(&DBOutbound{}).
		Where("tag = ? AND source = ?", outbound.Tag, entity.OutboundSourceManual).
		Updates(map[string]any{
			"name":            outbound.Name,
			"description":     outbound.Description,
			"type":            outbound.Type,
			"enabled":         outbound.Enabled,
			"sort":            outbound.Sort,
			"visible_devices": outbound.VisibleDevices,
			"config_json":     outbound.ConfigJSON,
			"source":          outbound.Source,
			"subscribe_name":  outbound.SubscribeName,
			"last_fetch_time": outbound.LastFetchTime,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ds *DatabaseStorage) DeleteExtraOutbound(tag string) error {
	result := ds.db.Where("tag = ? AND source = ?", tag, entity.OutboundSourceManual).Delete(&DBOutbound{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListOutbounds 按条件查询统一 Outbound 表。
func (ds *DatabaseStorage) GetOutbound(id int64) (*entity.Outbound, error) {
	var item DBOutbound
	if err := ds.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return item.ToEntity(), nil
}

// ListOutbounds 按条件查询统一 Outbound 表。
func (ds *DatabaseStorage) ListOutbounds(filters ...OutboundFilter) ([]*entity.Outbound, error) {
	query := ds.db.Model(&DBOutbound{})
	for _, filter := range filters {
		if filter.Source != nil {
			query = query.Where("source = ?", *filter.Source)
		}
		if filter.SubscribeName != "" {
			query = query.Where("subscribe_name = ?", filter.SubscribeName)
		}
		if filter.Enabled != nil {
			query = query.Where("enabled = ?", *filter.Enabled)
		}
	}

	var dbOutbounds []DBOutbound
	if err := query.Find(&dbOutbounds).Error; err != nil {
		return nil, err
	}

	result := make([]*entity.Outbound, 0, len(dbOutbounds))
	for _, item := range dbOutbounds {
		result = append(result, item.ToEntity())
	}
	return result, nil
}

// GetOutboundsByDevice 返回某个设备可见且启用的 Outbound。
func (ds *DatabaseStorage) GetOutboundsByDevice(deviceCode string) ([]*entity.Outbound, error) {
	enabled := true
	outbounds, err := ds.ListOutbounds(OutboundFilter{Enabled: &enabled})
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Outbound, 0, len(outbounds))
	for _, outbound := range outbounds {
		if outbound == nil || !databaseOutboundVisible(outbound.VisibleDevices, deviceCode) {
			continue
		}
		result = append(result, outbound)
	}
	return result, nil
}

// UpdateOutbound 按主键更新统一 Outbound 记录。
func (ds *DatabaseStorage) UpdateOutbound(outbound *entity.Outbound) error {
	if outbound == nil || outbound.ID == 0 {
		return ErrNotFound
	}
	result := ds.db.Model(&DBOutbound{}).Where("id = ?", outbound.ID).Updates(databaseOutboundValues(outbound))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteOutbound 按主键删除统一 Outbound 记录。
func (ds *DatabaseStorage) DeleteOutbound(id int64) error {
	result := ds.db.Delete(&DBOutbound{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateOrUpdateOutbounds 按 tag 批量覆盖写入 Outbound。
func (ds *DatabaseStorage) CreateOrUpdateOutbounds(items []*entity.Outbound) error {
	if len(items) == 0 {
		return nil
	}

	return ds.db.Transaction(func(tx *gorm.DB) error {
		for _, outbound := range items {
			if outbound == nil || outbound.Tag == "" {
				continue
			}
			if err := tx.Model(&DBOutbound{}).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "tag"}},
				DoUpdates: clause.Assignments(databaseOutboundValues(outbound)),
			}).Create(databaseOutboundValues(outbound)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func databaseOutboundValues(outbound *entity.Outbound) map[string]any {
	values := map[string]any{
		"tag":             outbound.Tag,
		"name":            outbound.Name,
		"description":     outbound.Description,
		"type":            outbound.Type,
		"enabled":         outbound.Enabled,
		"sort":            outbound.Sort,
		"visible_devices": outbound.VisibleDevices,
		"config_json":     outbound.ConfigJSON,
		"source":          outbound.Source,
		"subscribe_name":  outbound.SubscribeName,
		"last_fetch_time": outbound.LastFetchTime,
	}
	if outbound.ID != 0 {
		values["id"] = outbound.ID
	}
	return values
}

// DeleteOutboundsBySubscribe 删除指定订阅源下的缓存 Outbound。
func (ds *DatabaseStorage) DeleteOutboundsBySubscribe(subscribeName string) error {
	return ds.db.Where("source = ? AND subscribe_name = ?", entity.OutboundSourceSubscription, subscribeName).Delete(&DBOutbound{}).Error
}

// UpdateOutboundCacheTime 更新订阅的缓存时间与状态。
func (ds *DatabaseStorage) UpdateOutboundCacheTime(subscribeName string, timestamp time.Time) error {
	result := ds.db.Model(&DBSubscribe{}).Where("name = ?", subscribeName).Updates(map[string]any{
		"outbound_last_fetch_time":   timestamp,
		"outbound_last_fetch_status": "SUCCESS",
		"outbound_last_fetch_error":  "",
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func databaseOutboundVisible(visibleDevices, deviceCode string) bool {
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
