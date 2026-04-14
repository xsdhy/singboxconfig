import type { DeviceInbound } from '../types';

export interface DeviceInboundSelectionState {
  selectedTags: string[];
  sortByTag: Record<string, number>;
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
