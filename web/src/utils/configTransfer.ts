import type { ConfigImportSummary } from '../types';

/**
 * formatConfigImportSummary 把后端导入/种子初始化摘要整理成适合直接展示的中文文本。
 * 这样配置页和测试都可以复用同一份格式化逻辑，避免字段扩展后出现前后不一致。
 */
export function formatConfigImportSummary(summary: ConfigImportSummary): string {
  const lines = [
    `订阅: 导入 ${summary.subscribes.imported}，跳过 ${summary.subscribes.skipped}，失败 ${summary.subscribes.failed}`,
    `节点分组: 导入 ${summary.node_groups.imported}，跳过 ${summary.node_groups.skipped}，失败 ${summary.node_groups.failed}`,
    `规则集: 导入 ${summary.rule_sets.imported}，跳过 ${summary.rule_sets.skipped}，失败 ${summary.rule_sets.failed}`,
    `全局设置: 导入 ${summary.global_settings.imported}，跳过 ${summary.global_settings.skipped}，失败 ${summary.global_settings.failed}`,
    `设备: 导入 ${summary.devices.imported}，跳过 ${summary.devices.skipped}，失败 ${summary.devices.failed}`,
    `Inbound: 导入 ${summary.inbounds.imported}，跳过 ${summary.inbounds.skipped}，失败 ${summary.inbounds.failed}`,
    `设备 Inbound 绑定: 导入 ${summary.device_inbounds.imported}，跳过 ${summary.device_inbounds.skipped}，失败 ${summary.device_inbounds.failed}`,
    `WireGuard: 导入 ${summary.wire_guards.imported}，跳过 ${summary.wire_guards.skipped}，失败 ${summary.wire_guards.failed}`,
    `WireGuard Peer: 导入 ${summary.wire_guard_peers.imported}，跳过 ${summary.wire_guard_peers.skipped}，失败 ${summary.wire_guard_peers.failed}`,
    `额外出站: 导入 ${summary.extra_outbounds.imported}，跳过 ${summary.extra_outbounds.skipped}，失败 ${summary.extra_outbounds.failed}`,
  ];

  if (summary.errors.length > 0) {
    lines.push('', `错误明细: ${summary.errors.join('；')}`);
  }

  return lines.join('\n');
}
