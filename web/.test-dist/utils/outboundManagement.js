// buildOutboundQueryParams 统一把页面筛选状态映射成后端分页查询参数。
export function buildOutboundQueryParams(filters, page, limit) {
    const params = {
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
export function buildBatchEnablePayload(ids, enabled) {
    const normalizedIDs = Array.from(new Set(ids.filter((id) => typeof id === 'number' && id > 0)));
    return {
        ids: normalizedIDs,
        enabled,
    };
}
// createManualOutboundDraft 返回新增手工 Outbound 的默认草稿。
export function createManualOutboundDraft(sort) {
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
