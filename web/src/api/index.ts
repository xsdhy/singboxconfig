import axios from 'axios';
import type {
  Subscribe,
  NodeGroup,
  RuleSet,
  Setting,
  ConfigImportSummary,
  Device,
  Inbound,
  DeviceInbound,
  WireGuard,
  WireGuardPeer,
  Outbound,
  OutboundListResponse,
  RefreshSubscribeOutboundsResponse,
  SubscribeOutboundListResponse,
  AuthProfile,
  AuthLoginResponse,
  AuthChangeCredentialsResponse,
} from '../types';

const api = axios.create();
const AUTH_TOKEN_STORAGE_KEY = 'singboxconfig.auth.token';

api.interceptors.request.use((config) => {
  const token = window.localStorage.getItem(AUTH_TOKEN_STORAGE_KEY);
  if (token) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// encodePathSegment 统一编码路径参数，避免 tag/code 中的特殊字符破坏路由。
const encodePathSegment = (value: string | number) => encodeURIComponent(String(value));

// 订阅
export const getSubscribes = () => api.get<Subscribe[]>('/api/subscribes');
export const createSubscribe = (data: Subscribe) => api.post('/api/subscribes', data);
export const updateSubscribe = (name: string, data: Subscribe) => api.put(`/api/subscribes/${encodePathSegment(name)}`, data);
export const deleteSubscribe = (name: string) => api.delete(`/api/subscribes/${encodePathSegment(name)}`);
export const updateSubscribeCacheConfig = (name: string, outboundCacheDuration: number) =>
  api.put<Subscribe>(`/api/subscribes/${encodePathSegment(name)}/cache-config`, { outboundCacheDuration });
export const refreshSubscribeOutbounds = (name: string) =>
  api.post<RefreshSubscribeOutboundsResponse>(`/api/subscribes/${encodePathSegment(name)}/refresh-outbound`);
export const getSubscribeOutbounds = (name: string, page = 1, limit = 10) =>
  api.get<SubscribeOutboundListResponse>(`/api/subscribes/${encodePathSegment(name)}/outbounds`, {
    params: { page, limit },
  });

// 节点分组
export const getNodeGroups = () => api.get<NodeGroup[]>('/api/node-groups');
export const createNodeGroup = (data: NodeGroup) => api.post('/api/node-groups', data);
export const updateNodeGroup = (tag: string, data: NodeGroup) => api.put(`/api/node-groups/${encodePathSegment(tag)}`, data);
export const deleteNodeGroup = (tag: string) => api.delete(`/api/node-groups/${encodePathSegment(tag)}`);

// 规则集
export const getRuleSets = () => api.get<RuleSet[]>('/api/rule-sets');
export const createRuleSet = (data: RuleSet) => api.post('/api/rule-sets', data);
export const updateRuleSet = (tag: string, data: RuleSet) => api.put(`/api/rule-sets/${encodePathSegment(tag)}`, data);
export const deleteRuleSet = (tag: string) => api.delete(`/api/rule-sets/${encodePathSegment(tag)}`);

// 全局设置
export const getSettings = () => api.get<Setting[]>('/api/settings');
export const createSetting = (data: Setting) => api.post('/api/settings', data);
export const updateSetting = (key: string, data: Setting) => api.put(`/api/settings/${encodePathSegment(key)}`, data);
export const deleteSetting = (key: string) => api.delete(`/api/settings/${encodePathSegment(key)}`);
export const getSettingByKey = (key: string) => api.get<Setting>(`/api/settings/key/${encodePathSegment(key)}`);

// 认证管理
export const getStoredAuthToken = () => window.localStorage.getItem(AUTH_TOKEN_STORAGE_KEY);
export const setStoredAuthToken = (token: string) => window.localStorage.setItem(AUTH_TOKEN_STORAGE_KEY, token);
export const clearStoredAuthToken = () => window.localStorage.removeItem(AUTH_TOKEN_STORAGE_KEY);
export const login = (username: string, password: string) =>
  api.post<AuthLoginResponse>('/api/auth/login', { username, password });
export const getAuthMe = () => api.get<AuthProfile>('/api/auth/me');
export const changeCredentials = (oldPassword: string, newUsername?: string, newPassword?: string) =>
  api.post<AuthChangeCredentialsResponse>('/api/auth/change-credentials', {
    old_password: oldPassword,
    new_username: newUsername,
    new_password: newPassword,
  });

// 设备管理：设备本身和它与 Inbound 的绑定关系拆成两组接口。
export const getDevices = () => api.get<Device[]>('/api/devices');
export const createDevice = (data: Device) => api.post('/api/devices', data);
export const updateDevice = (code: string, data: Device) => api.put(`/api/devices/${encodePathSegment(code)}`, data);
export const deleteDevice = (code: string) => api.delete(`/api/devices/${encodePathSegment(code)}`);
export const getDeviceInbounds = (code: string) => api.get<DeviceInbound[]>(`/api/devices/${encodePathSegment(code)}/inbounds`);
export const setDeviceInbounds = (code: string, data: DeviceInbound[]) =>
  api.put(`/api/devices/${encodePathSegment(code)}/inbounds`, data);

// Inbound 模板
export const getInbounds = () => api.get<Inbound[]>('/api/inbounds');
export const createInbound = (data: Inbound) => api.post('/api/inbounds', data);
export const updateInbound = (tag: string, data: Inbound) => api.put(`/api/inbounds/${encodePathSegment(tag)}`, data);
export const deleteInbound = (tag: string) => api.delete(`/api/inbounds/${encodePathSegment(tag)}`);

// WireGuard 模板与 Peer
export const getWireGuards = () => api.get<WireGuard[]>('/api/wire-guards');
export const createWireGuard = (data: WireGuard) => api.post('/api/wire-guards', data);
export const updateWireGuard = (tag: string, data: WireGuard) => api.put(`/api/wire-guards/${encodePathSegment(tag)}`, data);
export const deleteWireGuard = (tag: string) => api.delete(`/api/wire-guards/${encodePathSegment(tag)}`);
export const getWireGuardPeers = (tag: string) => api.get<WireGuardPeer[]>(`/api/wire-guards/${encodePathSegment(tag)}/peers`);
export const createWireGuardPeer = (tag: string, data: WireGuardPeer) => api.post(`/api/wire-guards/${encodePathSegment(tag)}/peers`, data);
export const updateWireGuardPeer = (tag: string, id: number, data: WireGuardPeer) =>
  api.put(`/api/wire-guards/${encodePathSegment(tag)}/peers/${encodePathSegment(id)}`, data);
export const deleteWireGuardPeer = (tag: string, id: number) => api.delete(`/api/wire-guards/${encodePathSegment(tag)}/peers/${encodePathSegment(id)}`);

// 统一 Outbound：同时覆盖手工出站和订阅缓存节点。
export const getOutbounds = (params?: Record<string, string | number | boolean | undefined>) =>
  api.get<OutboundListResponse>('/api/outbounds', { params });
export const createOutbound = (data: Outbound) => api.post<Outbound>('/api/outbounds', data);
export const updateOutbound = (id: number, data: Outbound) => api.put<Outbound>(`/api/outbounds/${encodePathSegment(id)}`, data);
export const deleteOutbound = (id: number) => api.delete(`/api/outbounds/${encodePathSegment(id)}`);
export const batchEnableOutbounds = (ids: number[], enabled: boolean) =>
  api.patch<{ updated: number; enabled: boolean }>('/api/outbounds/batch-enable', { ids, enabled });

// 导入导出接口用于整库迁移和备份恢复。
export const exportConfig = () =>
  api.get<Blob>('/api/config-transfer/export', { responseType: 'blob' });

export const importConfig = (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  return api.post<ConfigImportSummary>('/api/config-transfer/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};
