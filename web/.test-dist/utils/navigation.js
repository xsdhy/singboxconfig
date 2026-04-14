export const MANAGEMENT_NAV_ITEMS = [
    { key: 'devices', title: 'Devices', description: '设备身份、WireGuard 绑定与 Inbound 勾选' },
    { key: 'dns', title: 'DNS', description: '独立维护全局 DNS JSON' },
    { key: 'inbounds', title: 'Inbound', description: '维护 Inbound 模板与 JSON 配置' },
    { key: 'subscribes', title: 'Subscribes', description: '维护订阅提供商与链接信息' },
    { key: 'node-groups', title: 'Outbound Groups', description: '定义节点筛选与测速逻辑' },
    { key: 'outbounds', title: 'Outbound', description: '统一维护手工 Outbound 与订阅缓存节点' },
    { key: 'rule-sets', title: 'Rules', description: '配置远程或本地分流规则' },
    { key: 'wireguards', title: 'WireGuard', description: '维护 WireGuard 模板与 Peer 明细' },
    { key: 'settings', title: 'Global Settings', description: '管理核心全局 Key-Value 设置' },
];
