import type { Subscribe, SubscribeCacheInfo } from '../types';

export interface SubscribeCacheSummary {
  label: string;
  detail: string;
  color: 'green' | 'orange' | 'red' | 'gray';
}

function formatDateTime(value?: string | null) {
  if (!value) {
    return '未拉取';
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }

  return parsed.toLocaleString('zh-CN', {
    hour12: false,
  });
}

// summarizeSubscribeCache 把订阅缓存元数据转换成页面展示文案，避免组件里散落判断分支。
export function summarizeSubscribeCache(
  subscribe: Pick<Subscribe, 'outboundLastFetchTime' | 'outboundCacheDuration' | 'outboundLastFetchStatus' | 'outboundLastFetchError'>,
  now = new Date(),
): SubscribeCacheSummary {
  const cacheDuration = subscribe.outboundCacheDuration ?? 0;
  const lastFetchTime = subscribe.outboundLastFetchTime;
  const lastFetchStatus = subscribe.outboundLastFetchStatus ?? '';

  if (!lastFetchTime) {
    return {
      label: '未拉取',
      detail: '',
      color: 'gray',
    };
  }

  if (lastFetchStatus === 'FAILED') {
    return {
      label: '最近失败',
      detail: subscribe.outboundLastFetchError?.trim() || `上次尝试时间：${formatDateTime(lastFetchTime)}`,
      color: 'red',
    };
  }

  if (cacheDuration <= 0) {
    return {
      label: '不缓存',
      detail: `最近一次成功拉取：${formatDateTime(lastFetchTime)}`,
      color: 'orange',
    };
  }

  const expireAt = new Date(lastFetchTime).getTime() + cacheDuration * 60 * 1000;
  if (Number.isNaN(expireAt) || expireAt <= now.getTime()) {
    return {
      label: '已过期',
      detail: `最近一次成功拉取：${formatDateTime(lastFetchTime)}`,
      color: 'red',
    };
  }

  const remainMinutes = Math.max(1, Math.ceil((expireAt - now.getTime()) / (60 * 1000)));
  return {
    label: '有效',
    detail: `最近一次成功拉取：${formatDateTime(lastFetchTime)}，剩余约 ${remainMinutes} 分钟`,
    color: 'green',
  };
}

// summarizeDrawerCacheInfo 把订阅节点抽屉里的缓存摘要组织成统一展示格式。
export function summarizeDrawerCacheInfo(cacheInfo: SubscribeCacheInfo, now = new Date()) {
  return summarizeSubscribeCache(
    {
      outboundLastFetchTime: cacheInfo.lastFetchTime,
      outboundCacheDuration: cacheInfo.cacheDuration,
      outboundLastFetchStatus: cacheInfo.isExpired ? 'FAILED' : 'SUCCESS',
      outboundLastFetchError: cacheInfo.isExpired ? '当前缓存已过期' : '',
    },
    now,
  );
}
