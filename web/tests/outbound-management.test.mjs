import test from 'node:test';
import assert from 'node:assert/strict';
import {
  buildBatchEnablePayload,
  buildOutboundQueryParams,
  createManualOutboundDraft,
} from '../.test-dist/utils/outboundManagement.js';

test('buildOutboundQueryParams 会忽略空筛选并保留分页参数', () => {
  const params = buildOutboundQueryParams({
    source: 'ALL',
    enabled: 'ALL',
    subscribeName: '  ',
    search: '  hk  ',
  }, 2, 20);

  assert.deepEqual(params, {
    page: 2,
    limit: 20,
    search: 'hk',
  });
});

test('buildOutboundQueryParams 会映射来源、状态和订阅名', () => {
  const params = buildOutboundQueryParams({
    source: 'SUBSCRIPTION',
    enabled: 'false',
    subscribeName: 'sub-a',
    search: '',
  }, 1, 12);

  assert.deepEqual(params, {
    page: 1,
    limit: 12,
    source: 'SUBSCRIPTION',
    enabled: false,
    subscribe_name: 'sub-a',
  });
});

test('buildBatchEnablePayload 会去重并过滤非法 ID', () => {
  const payload = buildBatchEnablePayload([3, undefined, 2, 3, 0, -1], true);

  assert.deepEqual(payload, {
    ids: [3, 2],
    enabled: true,
  });
});

test('createManualOutboundDraft 返回可直接用于新增弹窗的默认值', () => {
  const draft = createManualOutboundDraft(8);

  assert.equal(draft.source, 'MANUAL');
  assert.equal(draft.subscribeName, '');
  assert.equal(draft.enabled, true);
  assert.equal(draft.sort, 8);
  assert.equal(draft.configJson, '{}');
});
