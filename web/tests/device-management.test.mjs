import test from 'node:test';
import assert from 'node:assert/strict';
import {
  buildDeviceInboundPayload,
  buildDeviceInboundSelection,
} from '../.test-dist/utils/deviceManagement.js';

test('buildDeviceInboundSelection 能提取勾选项和排序映射', () => {
  const state = buildDeviceInboundSelection([
    { deviceCode: 'phone', inboundTag: 'tun', sort: 20 },
    { deviceCode: 'phone', inboundTag: 'mixed', sort: 10 },
  ]);

  assert.deepEqual(state.selectedTags, ['tun', 'mixed']);
  assert.deepEqual(state.sortByTag, { tun: 20, mixed: 10 });
});

test('buildDeviceInboundPayload 会按排序和 tag 生成稳定 payload', () => {
  const payload = buildDeviceInboundPayload(
    'phone',
    ['mixed', 'tun', 'http'],
    { tun: 20, mixed: 10, http: 10 },
  );

  assert.deepEqual(payload, [
    { deviceCode: 'phone', inboundTag: 'http', sort: 10 },
    { deviceCode: 'phone', inboundTag: 'mixed', sort: 10 },
    { deviceCode: 'phone', inboundTag: 'tun', sort: 20 },
  ]);
});
