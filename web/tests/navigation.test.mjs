import test from 'node:test';
import assert from 'node:assert/strict';
import { MANAGEMENT_NAV_ITEMS } from '../.test-dist/utils/navigation.js';

test('导航项覆盖当前管理台的全部入口，并包含统一 Outbound 页面', () => {
  const keys = MANAGEMENT_NAV_ITEMS.map((item) => item.key);
  assert.deepEqual(keys, [
    'devices',
    'dns',
    'inbounds',
    'subscribes',
    'node-groups',
    'outbounds',
    'rule-sets',
    'wireguards',
    'settings',
  ]);
});
