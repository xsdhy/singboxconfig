import test from 'node:test';
import assert from 'node:assert/strict';
import { buildDeleteConfirmContent } from '../.test-dist/utils/deleteConfirm.js';

test('buildDeleteConfirmContent 会拼接资源名、标识和影响说明', () => {
  const text = buildDeleteConfirmContent('WireGuard', 'wg-main', '相关 Peer 和设备绑定会同步清除。');

  assert.match(text, /确认删除WireGuard“wg-main”吗？/);
  assert.match(text, /相关 Peer 和设备绑定会同步清除。/);
  assert.match(text, /该操作不可撤销。/);
});

test('buildDeleteConfirmContent 在没有影响说明时仍保留不可撤销提示', () => {
  const text = buildDeleteConfirmContent('设备', 'phone');

  assert.equal(text, '确认删除设备“phone”吗？ 该操作不可撤销。');
});
