import test from 'node:test';
import assert from 'node:assert/strict';
import { summarizeSubscribeCache } from '../.test-dist/utils/subscribeOutbound.js';

test('summarizeSubscribeCache 在没有拉取记录时返回未拉取状态', () => {
  const summary = summarizeSubscribeCache({
    outboundLastFetchTime: null,
    outboundCacheDuration: 30,
    outboundLastFetchStatus: '',
    outboundLastFetchError: '',
  });

  assert.equal(summary.label, '未拉取');
  assert.equal(summary.color, 'gray');
});

test('summarizeSubscribeCache 在失败状态时优先显示错误', () => {
  const summary = summarizeSubscribeCache({
    outboundLastFetchTime: '2026-04-13T08:00:00Z',
    outboundCacheDuration: 30,
    outboundLastFetchStatus: 'FAILED',
    outboundLastFetchError: 'timeout',
  });

  assert.equal(summary.label, '最近失败');
  assert.equal(summary.detail, 'timeout');
  assert.equal(summary.color, 'red');
});

test('summarizeSubscribeCache 能区分有效缓存和过期缓存', () => {
  const valid = summarizeSubscribeCache({
    outboundLastFetchTime: '2026-04-13T08:40:00Z',
    outboundCacheDuration: 30,
    outboundLastFetchStatus: 'SUCCESS',
    outboundLastFetchError: '',
  }, new Date('2026-04-13T09:00:00Z'));
  const expired = summarizeSubscribeCache({
    outboundLastFetchTime: '2026-04-13T08:20:00Z',
    outboundCacheDuration: 30,
    outboundLastFetchStatus: 'SUCCESS',
    outboundLastFetchError: '',
  }, new Date('2026-04-13T09:00:00Z'));

  assert.equal(valid.label, '有效');
  assert.equal(valid.color, 'green');
  assert.match(valid.detail, /剩余约 10 分钟/);

  assert.equal(expired.label, '已过期');
  assert.equal(expired.color, 'red');
});
