import test from 'node:test';
import assert from 'node:assert/strict';
import {
  buildDeviceConfigURL,
  buildDeviceInboundPayload,
  buildDeviceInboundSelection,
  DEVICE_CONFIG_OUTPUTS,
} from '../.test-dist/utils/deviceManagement.js';

test('buildDeviceConfigURL 能为不同客户端生成公开配置链接', () => {
  const device = { code: 'phone main', token: 'token-123' };
  const origin = 'http://localhost:7391';

  assert.equal(
    buildDeviceConfigURL(device, 'singbox', origin),
    'http://localhost:7391/open/generate/phone%20main?token=token-123',
  );
  assert.equal(
    buildDeviceConfigURL(device, 'surge', origin),
    'http://localhost:7391/open/surge/phone%20main?token=token-123',
  );
  assert.equal(
    buildDeviceConfigURL(device, 'shadowrocket', origin),
    'http://localhost:7391/open/shadowrocket/phone%20main?token=token-123',
  );
  assert.equal(DEVICE_CONFIG_OUTPUTS.shadowrocket.label, 'Shadowrocket');
});

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
