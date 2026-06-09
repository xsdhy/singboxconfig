import type { DeviceInbound } from '../types';

export type DeviceConfigOutputType = 'singbox' | 'surge' | 'shadowrocket';

export interface DeviceConfigOutputMeta {
  endpoint: string;
  label: string;
}

export const DEVICE_CONFIG_OUTPUTS: Record<DeviceConfigOutputType, DeviceConfigOutputMeta> = {
  singbox: { endpoint: 'generate', label: 'sing-box' },
  surge: { endpoint: 'surge', label: 'Surge' },
  shadowrocket: { endpoint: 'shadowrocket', label: 'Shadowrocket' },
};

export interface DeviceInboundSelectionState {
  selectedTags: string[];
  sortByTag: Record<string, number>;
}

/**
 * buildDeviceConfigURL 根据设备和客户端输出类型构造公开订阅链接。
 */
export function buildDeviceConfigURL(record: { code: string; token: string }, outputType: DeviceConfigOutputType, origin: string): string {
  const output = DEVICE_CONFIG_OUTPUTS[outputType];
  const url = new URL(`/open/${output.endpoint}/${encodeURIComponent(record.code)}`, origin);
  url.searchParams.set('token', record.token);
  return url.toString();
}

/**
 * buildDeviceInboundSelection 把后端绑定列表转换成前端表单可直接使用的状态。
 */
export function buildDeviceInboundSelection(bindings: DeviceInbound[]): DeviceInboundSelectionState {
  const selectedTags: string[] = [];
  const sortByTag: Record<string, number> = {};

  bindings.forEach((binding) => {
    selectedTags.push(binding.inboundTag);
    sortByTag[binding.inboundTag] = binding.sort;
  });

  return { selectedTags, sortByTag };
}

/**
 * buildDeviceInboundPayload 根据勾选结果生成提交给后端的全量替换 payload。
 */
export function buildDeviceInboundPayload(
  deviceCode: string,
  selectedTags: string[],
  sortByTag: Record<string, number>,
): DeviceInbound[] {
  return [...selectedTags]
    .sort((left, right) => {
      const leftSort = sortByTag[left] ?? 0;
      const rightSort = sortByTag[right] ?? 0;
      if (leftSort === rightSort) {
        return left.localeCompare(right);
      }
      return leftSort - rightSort;
    })
    .map((inboundTag) => ({
      deviceCode,
      inboundTag,
      sort: sortByTag[inboundTag] ?? 0,
    }));
}
