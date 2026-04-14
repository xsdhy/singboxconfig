import type { Outbound, OutboundSource } from '../types';

export interface OutboundFilters {
  source: 'ALL' | OutboundSource;
  enabled: 'ALL' | 'true' | 'false';
  subscribeName: string;
  search: string;
}

export interface OutboundQueryParams {
  [key: string]: string | number | boolean | undefined;
  source?: OutboundSource;
  enabled?: boolean;
  subscribe_name?: string;
  search?: string;
  page: number;
  limit: number;
}

// buildOutboundQueryParams 统一把页面筛选状态映射成后端分页查询参数。
export function buildOutboundQueryParams(filters: OutboundFilters, page: number, limit: number): OutboundQueryParams {
  const params: OutboundQueryParams = {
    page,
    limit,
  };

  if (filters.source !== 'ALL') {
    params.source = filters.source;
  }
  if (filters.enabled !== 'ALL') {
    params.enabled = filters.enabled === 'true';
  }
  if (filters.subscribeName.trim()) {
    params.subscribe_name = filters.subscribeName.trim();
  }
  if (filters.search.trim()) {
    params.search = filters.search.trim();
  }

  return params;
}

// buildBatchEnablePayload 会去重并过滤非法 ID，避免前端误发空请求。
export function buildBatchEnablePayload(ids: Array<number | undefined>, enabled: boolean) {
  const normalizedIDs = Array.from(
    new Set(ids.filter((id): id is number => typeof id === 'number' && id > 0)),
  );

  return {
    ids: normalizedIDs,
    enabled,
  };
}

// createManualOutboundDraft 返回新增手工 Outbound 的默认草稿。
export function createManualOutboundDraft(sort: number): Outbound {
  return {
    tag: '',
    name: '',
    description: '',
    type: '',
    enabled: true,
    sort,
    visibleDevices: '',
    configJson: '{}',
    source: 'MANUAL',
    subscribeName: '',
  };
}
