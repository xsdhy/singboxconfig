import test from 'node:test';
import assert from 'node:assert/strict';
import {
  DEFAULT_DNS_CONFIG,
  DNS_SETTING_KEY,
  getDefaultDnsEditorText,
  getDnsEditorText,
} from '../.test-dist/utils/dns.js';

test('DNS 常量使用约定的设置 key', () => {
  assert.equal(DNS_SETTING_KEY, 'dns_config');
});

test('getDefaultDnsEditorText 会输出默认 DNS 的格式化 JSON', () => {
  const text = getDefaultDnsEditorText();
  assert.match(text, /"final": "dns_proxy"/);
  assert.match(text, /"tag": "dns_proxy"/);
});

test('getDnsEditorText 在没有设置值时回退到默认模板', () => {
  assert.equal(getDnsEditorText(undefined), getDefaultDnsEditorText());
});

test('默认 DNS 模板包含代理和直连两个核心服务器', () => {
  const tags = DEFAULT_DNS_CONFIG.servers.map((server) => server.tag);
  assert.deepEqual(tags.slice(0, 2), ['dns_proxy', 'dns_direct']);
});
