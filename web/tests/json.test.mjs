import test from 'node:test';
import assert from 'node:assert/strict';
import { normalizeJsonText, prettyJsonText } from '../.test-dist/utils/json.js';

test('prettyJsonText 会把 JSON 字符串格式化为多行文本', () => {
  const text = prettyJsonText('{"name":"phone","enabled":true}');
  assert.equal(text, '{\n  "name": "phone",\n  "enabled": true\n}');
});

test('prettyJsonText 在空值场景下回退到默认文本', () => {
  assert.equal(prettyJsonText('', '{ "ok": true }'), '{ "ok": true }');
});

test('normalizeJsonText 会校验并压缩 JSON 文本', () => {
  const normalized = normalizeJsonText('{\n  "servers": [1, 2]\n}');
  assert.equal(normalized, '{"servers":[1,2]}');
});
