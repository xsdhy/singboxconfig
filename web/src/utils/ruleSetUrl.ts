import type { Device } from '../types';

// 规则集 open 接口支持的软件，与后端 entity.ParseSoftware / convert/ruleset 的取值保持一致。
export type RuleSetSoftware = 'singbox' | 'surge' | 'shadowrocket';

export const RULE_SET_SOFTWARES: { key: RuleSetSoftware; label: string }[] = [
  { key: 'singbox', label: 'sing-box' },
  { key: 'surge', label: 'Surge' },
  { key: 'shadowrocket', label: 'Shadowrocket' },
];

// buildRuleSetOpenURL 拼接规则集 open 接口的绝对访问地址：
// <base>/open/rules/<tag>/<software>/<device>?token=<token>
// 与后端 convert/ruleset.BuildRuleSetURL 的转义规则一致：tag、device 走 path 转义，token 走 query 转义。
export function buildRuleSetOpenURL(base: string, tag: string, software: RuleSetSoftware, device: Device): string {
  const host = base.trim().replace(/\/+$/, '');
  return `${host}/open/rules/${encodeURIComponent(tag)}/${software}/${encodeURIComponent(device.code)}?token=${encodeURIComponent(device.token)}`;
}

// isRuleSetVisibleToDevice 复用后端可见性判定：AbleDevices 为空表示全部可见，
// 非空时沿用 strings.Contains(ableDevices, deviceCode) 逻辑。
export function isRuleSetVisibleToDevice(ableDevices: string | undefined, deviceCode: string): boolean {
  const able = (ableDevices || '').trim();
  if (able === '') {
    return true;
  }
  return able.includes(deviceCode);
}
