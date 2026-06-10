// Subscribe 对应后端订阅源实体。
export interface Subscribe {
  name: string; // 订阅名称，前端列表里也把它当唯一键使用。
  url: string; // 远程订阅地址。
  userAgent?: string; // 可选请求头，留空时后端会走默认值。
  status: boolean; // 是否参与节点抓取和配置生成。
  visibleDevices?: string; // 留空表示所有设备都可见。
  outboundLastFetchTime?: string | null; // 最近一次成功刷新缓存的时间。
  outboundCacheDuration?: number; // 缓存时长，单位分钟。
  outboundLastFetchStatus?: string; // SUCCESS / FAILED / 空字符串。
  outboundLastFetchError?: string; // 最近一次失败原因。
}

// NodeGroup 描述节点筛选与分组规则。
export interface NodeGroup {
  id?: number;
  name: string;
  tag: string; // 分组唯一标识，会被规则集等配置引用。
  groupType: string; // selector / urltest。
  include?: string; // 逗号分隔关键字。
  exclude?: string; // 逗号分隔关键字。
  testURL?: string; // urltest 可用性测试地址。
  // deviceTypeOverrides 设备级分组类型覆盖规则，逗号分隔的 "设备编码:分组类型" 列表，
  // 例如 "phone:selector,gateway:urltest"。未配置覆盖的设备使用 groupType 默认类型。
  deviceTypeOverrides?: string;
}

// RuleSet 描述本地或远程规则集。
export interface RuleSet {
  id?: number;
  name: string;
  tag: string; // 规则集唯一标识。
  ruleSetType: string; // local / remote。
  format: string; // source / binary。
  url?: string;
  downloadDetour?: string;
  ableDevices?: string;
  content?: string | object; // local 规则集在编辑时可能是对象，提交时会转回字符串。
  outbound?: string; // 命中该规则集后的默认出站。
  sort: number; // 路由规则顺序。
}

// Setting 是简单的全局 key/value 配置。
export interface Setting {
  id?: number;
  key: string;
  value: string;
}

// Device 描述配置生成的目标设备。
export interface Device {
  code: string; // 设备编码，也是生成接口路径参数。
  name: string;
  description: string;
  token: string; // 拉取配置时使用的鉴权 token。
  enabled: boolean;
  sort: number;
  wireGuardTag: string; // 绑定的 WireGuard 模板 tag。
  wireGuardClientAddr: string; // 当前设备的 WireGuard 地址。
  wireGuardClientKey: string; // 当前设备的 WireGuard 私钥。
}

// Inbound 表示一份可复用的入站模板。
export interface Inbound {
  tag: string;
  name: string;
  description: string;
  type: string;
  enabled: boolean;
  sort: number;
  configJson: string; // 原始 sing-box inbound JSON。
}

// DeviceInbound 表示设备和入站模板的绑定关系。
export interface DeviceInbound {
  deviceCode: string;
  inboundTag: string;
  sort: number;
}

// WireGuard 描述一份 WireGuard endpoint 模板。
export interface WireGuard {
  tag: string;
  name: string;
  description: string;
  enabled: boolean;
  sort: number;
  endpointTag: string; // 输出到 sing-box endpoint.tag。
  mtu: number;
}

// WireGuardPeer 表示 WireGuard 模板下的单个对端。
export interface WireGuardPeer {
  id?: number;
  wireGuardTag: string;
  sort: number;
  address: string;
  port: number;
  publicKey: string;
  preSharedKey: string;
  allowedIps: string;
  persistentKeepaliveInterval: number;
  enabled: boolean;
}

// OutboundSource 区分手工维护和订阅缓存两类出站。
export type OutboundSource = 'MANUAL' | 'SUBSCRIPTION';

// Outbound 表示统一出站实体。
export interface Outbound {
  id?: number;
  tag: string;
  name: string;
  description: string;
  type: string;
  enabled: boolean;
  sort: number;
  visibleDevices: string;
  configJson: string;
  source: OutboundSource;
  subscribeName: string;
  lastFetchTime?: string | null;
  createdAt?: string;
  updatedAt?: string;
}

// OutboundListResponse 是统一 Outbound 列表接口的分页返回体。
export interface OutboundListResponse {
  items: Outbound[];
  total: number;
  page: number;
  limit: number;
}

// SubscribeCacheInfo 是订阅缓存摘要。
export interface SubscribeCacheInfo {
  lastFetchTime?: string | null;
  cacheDuration: number;
  isExpired: boolean;
}

// SubscribeOutboundListResponse 是单个订阅缓存节点的分页返回体。
export interface SubscribeOutboundListResponse extends OutboundListResponse {
  subscribeCacheInfo: SubscribeCacheInfo;
}

// RefreshSubscribeOutboundsResponse 描述手动刷新后的变更统计。
export interface RefreshSubscribeOutboundsResponse {
  status: string;
  added: number;
  updated: number;
  deleted: number;
  lastFetchTime?: string | null;
}

// SingDNSServer / SingDNSRule / SingDNS 对应 DNS JSON 编辑器里的结构。
export interface SingDNSServer {
  tag: string;
  address: string;
  address_resolver?: string;
  strategy?: string;
  detour?: string;
}

export interface SingDNSRule {
  outbound?: string;
  server?: string;
  clash_mode?: string;
  rule_set?: string;
  protocol?: string;
  action?: string;
}

export interface SingDNS {
  servers: SingDNSServer[];
  rules: SingDNSRule[];
  final: string;
}

export interface NavigationItem {
  key: string;
  title: string;
  description: string;
}

export interface AuthProfile {
  username: string;
  initialized_at?: string;
  password_changed_at?: string;
}

export interface AuthLoginResponse {
  access_token: string;
  token_type: string;
  expires_at: string;
  username: string;
  password_changed_at?: string;
}

export interface AuthChangeCredentialsResponse extends AuthLoginResponse {
  message: string;
}

// ImportCategorySummary 记录某一类资源在导入时的处理结果。
export interface ImportCategorySummary {
  imported: number;
  skipped: number;
  failed: number;
}

// ConfigImportSummary 是配置导入和初始化默认数据接口的统一返回体。
export interface ConfigImportSummary {
  subscribes: ImportCategorySummary;
  node_groups: ImportCategorySummary;
  rule_sets: ImportCategorySummary;
  global_settings: ImportCategorySummary;
  devices: ImportCategorySummary;
  inbounds: ImportCategorySummary;
  device_inbounds: ImportCategorySummary;
  wire_guards: ImportCategorySummary;
  wire_guard_peers: ImportCategorySummary;
  extra_outbounds: ImportCategorySummary;
  errors: string[];
}
