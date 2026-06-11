import { useEffect, useMemo, useState } from 'react';
import { Modal, Select, Button, Typography, Message, Space, Empty } from '@arco-design/web-react';
import type { Device, RuleSet } from '../types';
import {
  RULE_SET_SOFTWARES,
  buildRuleSetOpenURL,
  isRuleSetVisibleToDevice,
  type RuleSetSoftware,
} from '../utils/ruleSetUrl';

const { Text } = Typography;

interface Props {
  visible: boolean;
  ruleSet: RuleSet | null;
  devices: Device[];
  systemHost: string;
  onCancel: () => void;
}

// RuleSetCopyURLModal 让用户选择一个设备后，按三种软件复制该规则集的 open 接口地址。
// 地址形如 <systemHost>/open/rules/:tag?software=...&device=...&token=...，必须带设备 + token。
export default function RuleSetCopyURLModal({ visible, ruleSet, devices, systemHost, onCancel }: Props) {
  // 只展示对当前规则集可见、且已启用的设备（禁用设备访问接口会被 403）。
  const visibleDevices = useMemo(
    () => devices.filter((d) => d.enabled && isRuleSetVisibleToDevice(ruleSet?.ableDevices, d.code)),
    [devices, ruleSet],
  );

  const [deviceCode, setDeviceCode] = useState('');

  useEffect(() => {
    if (visible) {
      setDeviceCode(visibleDevices[0]?.code || '');
    }
  }, [visible, visibleDevices]);

  const selectedDevice = visibleDevices.find((d) => d.code === deviceCode) || null;
  const hostMissing = systemHost.trim() === '';

  const handleCopy = async (software: RuleSetSoftware) => {
    if (!ruleSet || !selectedDevice) {
      return;
    }
    try {
      await navigator.clipboard.writeText(buildRuleSetOpenURL(systemHost, ruleSet.tag, software, selectedDevice));
      Message.success('地址已复制');
    } catch {
      Message.error('复制失败');
    }
  };

  return (
    <Modal
      visible={visible}
      title={ruleSet ? `复制规则集地址 · ${ruleSet.name}` : '复制规则集地址'}
      onCancel={onCancel}
      footer={null}
      style={{ width: 560 }}
    >
      {hostMissing && (
        <div style={{ marginBottom: 12 }}>
          <Text type="warning" style={{ fontSize: 12 }}>
            尚未在系统设置中配置「系统 Host」，复制出的地址将缺少前缀，请先到系统设置补全。
          </Text>
        </div>
      )}
      <div style={{ marginBottom: 16 }}>
        <Text type="secondary" style={{ fontSize: 12 }}>选择设备（用于鉴权 token）</Text>
        <Select
          value={deviceCode}
          onChange={setDeviceCode}
          style={{ width: '100%', marginTop: 4 }}
          placeholder="选择设备"
          showSearch
        >
          {visibleDevices.map((d) => (
            <Select.Option key={d.code} value={d.code}>
              {d.name} ({d.code})
            </Select.Option>
          ))}
        </Select>
      </div>

      {visibleDevices.length === 0 ? (
        <Empty description="没有可用于该规则集的已启用设备" />
      ) : (
        <Space direction="vertical" size={10} style={{ width: '100%' }}>
          {RULE_SET_SOFTWARES.map(({ key, label }) => {
            const url = ruleSet && selectedDevice ? buildRuleSetOpenURL(systemHost, ruleSet.tag, key, selectedDevice) : '';
            return (
              <div key={key} className="glass-card" style={{ padding: '10px 12px' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
                  <div style={{ minWidth: 0 }}>
                    <div style={{ fontWeight: 600, fontSize: 13 }}>{label}</div>
                    <div style={{ fontSize: 11, opacity: 0.6, marginTop: 2, wordBreak: 'break-all' }}>{url}</div>
                  </div>
                  <Button type="primary" size="small" disabled={!selectedDevice} onClick={() => handleCopy(key)}>
                    复制
                  </Button>
                </div>
              </div>
            );
          })}
        </Space>
      )}
    </Modal>
  );
}
