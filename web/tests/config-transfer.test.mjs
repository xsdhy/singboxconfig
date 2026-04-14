import test from 'node:test';
import assert from 'node:assert/strict';
import { formatConfigImportSummary } from '../.test-dist/utils/configTransfer.js';

test('导入摘要格式化覆盖阶段五新增实体', () => {
  const text = formatConfigImportSummary({
    subscribes: { imported: 1, skipped: 0, failed: 0 },
    node_groups: { imported: 2, skipped: 0, failed: 0 },
    rule_sets: { imported: 3, skipped: 0, failed: 0 },
    global_settings: { imported: 4, skipped: 0, failed: 0 },
    devices: { imported: 5, skipped: 1, failed: 0 },
    inbounds: { imported: 6, skipped: 2, failed: 0 },
    device_inbounds: { imported: 7, skipped: 3, failed: 0 },
    wire_guards: { imported: 8, skipped: 4, failed: 0 },
    wire_guard_peers: { imported: 9, skipped: 5, failed: 0 },
    extra_outbounds: { imported: 10, skipped: 6, failed: 0 },
    errors: ['bad peer'],
  });

  assert.match(text, /设备: 导入 5，跳过 1，失败 0/);
  assert.match(text, /Inbound: 导入 6，跳过 2，失败 0/);
  assert.match(text, /设备 Inbound 绑定: 导入 7，跳过 3，失败 0/);
  assert.match(text, /WireGuard: 导入 8，跳过 4，失败 0/);
  assert.match(text, /WireGuard Peer: 导入 9，跳过 5，失败 0/);
  assert.match(text, /额外出站: 导入 10，跳过 6，失败 0/);
  assert.match(text, /错误明细: bad peer/);
});
