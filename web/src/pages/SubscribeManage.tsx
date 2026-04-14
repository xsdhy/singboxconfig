import { useState, useEffect, useCallback } from 'react';
import {
  Button,
  Drawer,
  Message,
  Pagination,
  Space,
  Tag,
  Typography,
} from '@arco-design/web-react';
import SubscribeTable from '../components/SubscribeTable';
import SubscribeModal from '../components/SubscribeModal';
import PageToolbar from '../components/PageToolbar';
import DataState from '../components/DataState';
import * as api from '../api';
import type { Outbound, Subscribe, SubscribeCacheInfo } from '../types';
import { summarizeDrawerCacheInfo } from '../utils/subscribeOutbound';

function formatDisplayTime(value?: string | null) {
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

export default function SubscribeManage() {
  const [subscribes, setSubscribes] = useState<Subscribe[]>([]);
  const [drawerItems, setDrawerItems] = useState<Outbound[]>([]);
  const [drawerCacheInfo, setDrawerCacheInfo] = useState<SubscribeCacheInfo | null>(null);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [drawerTitle, setDrawerTitle] = useState('');
  const [drawerPage, setDrawerPage] = useState(1);
  const [drawerLimit] = useState(10);
  const [drawerTotal, setDrawerTotal] = useState(0);
  const [drawerLoading, setDrawerLoading] = useState(false);
  const [drawerRefreshing, setDrawerRefreshing] = useState(false);
  const [drawerSubscribeName, setDrawerSubscribeName] = useState('');
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [refreshingKey, setRefreshingKey] = useState<string | null>(null);
  const [viewingKey, setViewingKey] = useState<string | null>(null);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);
  const [subVisible, setSubVisible] = useState(false);
  const [subTitle, setSubTitle] = useState('添加订阅');
  const [subForm, setSubForm] = useState<Partial<Subscribe>>({});

  const loadData = useCallback(async (manual = false) => {
    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      const res = await api.getSubscribes();
      setSubscribes(Array.isArray(res.data) ? res.data : []);
    } catch {
      Message.error('加载订阅数据失败');
      setSubscribes([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  // loadSubscribeOutbounds 负责抽屉分页加载，也复用于刷新后的局部回显。
  const loadSubscribeOutbounds = useCallback(async (subscribeName: string, nextPage = 1, manual = false) => {
    try {
      if (manual) {
        setDrawerRefreshing(true);
      } else {
        setDrawerLoading(true);
      }

      const response = await api.getSubscribeOutbounds(subscribeName, nextPage, drawerLimit);
      setDrawerItems(Array.isArray(response.data.items) ? response.data.items : []);
      setDrawerCacheInfo(response.data.subscribeCacheInfo);
      setDrawerTotal(response.data.total ?? 0);
      setDrawerPage(response.data.page ?? nextPage);
    } catch {
      Message.error('加载订阅缓存节点失败');
      setDrawerItems([]);
      setDrawerCacheInfo(null);
      setDrawerTotal(0);
    } finally {
      setDrawerLoading(false);
      setDrawerRefreshing(false);
    }
  }, [drawerLimit]);

  const handleAddSub = () => {
    setSubForm({ status: true, outboundCacheDuration: 0, visibleDevices: '' });
    setSubTitle('添加订阅');
    setSubVisible(true);
  };
  const handleEditSub = (r: Subscribe) => { setSubForm({ ...r }); setSubTitle('编辑订阅'); setSubVisible(true); };
  const handleSubOk = async (values: Subscribe) => {
    try {
      setSubmitting(true);
      const payload: Subscribe = {
        ...subForm,
        ...values,
        name: values.name.trim(),
        url: values.url.trim(),
        userAgent: values.userAgent?.trim() || '',
        status: !!values.status,
        visibleDevices: values.visibleDevices?.trim() || '',
        outboundCacheDuration: Math.max(0, Math.trunc(values.outboundCacheDuration ?? 0)),
      };

      const originalName = subForm.name || payload.name;
      if (subForm.name) {
        await api.updateSubscribe(originalName, payload);
      } else {
        await api.createSubscribe(payload);
      }
      await api.updateSubscribeCacheConfig(payload.name, payload.outboundCacheDuration ?? 0);

      Message.success('保存成功');
      setSubVisible(false);
      await loadData();
      if (drawerVisible && drawerSubscribeName === originalName) {
        setDrawerSubscribeName(payload.name);
        setDrawerTitle(payload.name);
        await loadSubscribeOutbounds(payload.name, 1);
      }
    } catch {
      Message.error('保存失败');
    }
    finally { setSubmitting(false); }
  };
  const handleDeleteSub = async (r: Subscribe) => {
    try {
      setDeletingKey(r.name);
      await api.deleteSubscribe(r.name);
      Message.success('删除成功');
      await loadData();
    } catch { Message.error('删除失败'); }
    finally { setDeletingKey(null); }
  };

  const handleRefreshOutbounds = async (record: Subscribe) => {
    try {
      setRefreshingKey(record.name);
      const response = await api.refreshSubscribeOutbounds(record.name);
      const result = response.data;
      Message.success(`刷新完成：新增 ${result.added}，更新 ${result.updated}，删除 ${result.deleted}`);
      await loadData();
      if (drawerVisible && drawerSubscribeName === record.name) {
        await loadSubscribeOutbounds(record.name, 1);
      }
    } catch {
      Message.error('刷新订阅缓存失败');
    } finally {
      setRefreshingKey(null);
    }
  };

  const handleViewOutbounds = async (record: Subscribe) => {
    try {
      setViewingKey(record.name);
      setDrawerSubscribeName(record.name);
      setDrawerTitle(record.name);
      setDrawerVisible(true);
      await loadSubscribeOutbounds(record.name, 1);
    } finally {
      setViewingKey(null);
    }
  };

  const drawerSummary = drawerCacheInfo ? summarizeDrawerCacheInfo(drawerCacheInfo) : null;
  const drawerEnabledCount = drawerItems.filter((item) => item.enabled).length;
  const drawerDisabledCount = drawerItems.length - drawerEnabledCount;
  const drawerStats = [
    {
      label: '缓存时长',
      value: `${drawerCacheInfo?.cacheDuration ?? 0} 分钟`,
      detail: drawerCacheInfo?.cacheDuration ? '到期后需要重新拉取' : '每次读取均实时拉取',
    },
    {
      label: '最近拉取',
      value: formatDisplayTime(drawerCacheInfo?.lastFetchTime),
      detail: drawerCacheInfo?.isExpired ? '当前缓存已过期' : '缓存状态正常',
    },
    {
      label: '启用节点',
      value: `${drawerEnabledCount} / ${drawerItems.length}`,
      detail: `禁用 ${Math.max(0, drawerDisabledCount)} 个`,
    },
  ];
  const { Text } = Typography;

  return (
    <>
      <PageToolbar
        summary="订阅管理：维护订阅基础信息、缓存时长，并支持手动刷新订阅 Outbound。"
        count={subscribes.length}
        countLabel="个订阅"
        onRefresh={() => loadData(true)}
        refreshing={refreshing}
        onPrimaryAction={handleAddSub}
        primaryActionLabel="添加订阅"
      />

      <div style={{ paddingTop: 8 }}>
        <DataState
          loading={loading}
          isEmpty={subscribes.length === 0}
          emptyTitle="还没有订阅源"
          emptyDescription="先添加一个订阅地址，后续再逐步完善 UA、状态和同步策略。"
          createLabel="立即添加订阅"
          onCreate={handleAddSub}
        >
          <SubscribeTable
            data={subscribes}
            deletingKey={deletingKey}
            refreshingKey={refreshingKey}
            viewingKey={viewingKey}
            onEdit={handleEditSub}
            onDelete={handleDeleteSub}
            onRefreshOutbounds={handleRefreshOutbounds}
            onViewOutbounds={handleViewOutbounds}
          />
        </DataState>
      </div>

      <SubscribeModal visible={subVisible} title={subTitle} initialValues={subForm}
        confirmLoading={submitting}
        onOk={handleSubOk} onCancel={() => setSubVisible(false)} />

      <Drawer
        visible={drawerVisible}
        width={640}
        title={`Outbound 列表 · ${drawerTitle}`}
        bodyStyle={{ padding: 16 }}
        footer={(
          <Space>
            <Button loading={drawerRefreshing} onClick={() => void loadSubscribeOutbounds(drawerSubscribeName, drawerPage, true)}>
              立即刷新列表
            </Button>
            <Button type="primary" loading={refreshingKey === drawerSubscribeName} onClick={() => {
              const current = subscribes.find((item) => item.name === drawerSubscribeName);
              if (current) {
                void handleRefreshOutbounds(current);
              }
            }}
            >
              手动刷新缓存
            </Button>
          </Space>
        )}
        onCancel={() => {
          setDrawerVisible(false);
          setDrawerItems([]);
          setDrawerCacheInfo(null);
          setDrawerPage(1);
          setDrawerTotal(0);
          setDrawerSubscribeName('');
        }}
      >
        <div style={{ display: 'grid', gap: 16 }}>
          <div
            style={{
              padding: 18,
              borderRadius: 20,
              background: 'linear-gradient(145deg, rgba(255,255,255,0.96), rgba(245,248,255,0.88))',
              border: '1px solid rgba(255,255,255,0.7)',
              boxShadow: '0 12px 32px rgba(31, 38, 135, 0.08)',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 16, marginBottom: 16, flexWrap: 'wrap' }}>
              <div style={{ display: 'grid', gap: 6 }}>
                <Text type="secondary" style={{ fontSize: 12 }}>订阅缓存概览</Text>
                <div style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)', lineHeight: 1.2 }}>
                  {drawerTitle || '当前订阅'}
                </div>
                <Text type="secondary" style={{ fontSize: 13 }}>
                  {drawerSummary?.detail || '拉取后会显示缓存状态与有效期。'}
                </Text>
              </div>
              {drawerSummary && <Tag color={drawerSummary.color}>{drawerSummary.label}</Tag>}
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: 12 }}>
              {drawerStats.map((stat) => (
                <div
                  key={stat.label}
                  style={{
                    padding: '12px 14px',
                    borderRadius: 14,
                    background: 'rgba(255,255,255,0.68)',
                    border: '1px solid rgba(0,0,0,0.05)',
                  }}
                >
                  <div style={{ fontSize: 11, color: 'var(--text-secondary)', marginBottom: 6 }}>{stat.label}</div>
                  <div style={{ fontSize: 16, fontWeight: 700, color: 'var(--text-primary)' }}>{stat.value}</div>
                  <div style={{ fontSize: 12, color: 'var(--text-secondary)', marginTop: 4, lineHeight: 1.5 }}>{stat.detail}</div>
                </div>
              ))}
            </div>
          </div>

          <div className="glass-card" style={{ padding: 0, height: 'auto' }}>
            <div
              style={{
                padding: '14px 16px 12px',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                gap: 12,
                flexWrap: 'wrap',
                borderBottom: '1px solid var(--border-color)',
              }}
            >
              <div>
                <div style={{ fontSize: 15, fontWeight: 700, color: 'var(--text-primary)' }}>节点列表</div>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  按当前分页返回订阅缓存中的 Outbound 节点
                </Text>
              </div>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  flexWrap: 'wrap',
                }}
              >
                <Text type="secondary" style={{ fontSize: 12 }}>
                  当前页 {drawerItems.length} 条，累计 {drawerTotal} 条
                </Text>
              </div>
            </div>

            <div style={{ padding: 12 }}>
              <DataState
                loading={drawerLoading}
                isEmpty={drawerTotal === 0}
                emptyTitle="当前订阅还没有缓存节点"
                emptyDescription="可以先点击“手动刷新缓存”从订阅源拉取节点。"
              >
                <div style={{ display: 'grid', gap: 12 }}>
                  {drawerItems.map((item) => (
                    <div
                      key={item.id ?? item.tag}
                      style={{
                        padding: 14,
                        borderRadius: 14,
                        background: 'rgba(255,255,255,0.76)',
                        border: '1px solid rgba(0,0,0,0.05)',
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', gap: 14, alignItems: 'flex-start', flexWrap: 'wrap' }}>
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap', marginBottom: 8 }}>
                            <strong style={{ fontSize: 14, color: 'var(--text-primary)' }}>{item.name || item.tag}</strong>
                            <Tag color={item.enabled ? 'arcoblue' : 'gray'}>{item.enabled ? '启用' : '禁用'}</Tag>
                            <Tag>{item.type || 'unknown'}</Tag>
                          </div>
                          <div style={{ display: 'grid', gap: 6 }}>
                            <div style={{ fontSize: 12, color: 'var(--text-secondary)', wordBreak: 'break-all' }}>
                              Tag: <span style={{ color: 'var(--text-primary)' }}>{item.tag}</span>
                            </div>
                            <div style={{ fontSize: 12, color: 'var(--text-secondary)', wordBreak: 'break-all' }}>
                              订阅源: <span style={{ color: 'var(--text-primary)' }}>{item.subscribeName || drawerSubscribeName || '-'}</span>
                            </div>
                            {item.description ? (
                              <div style={{ fontSize: 12, color: 'var(--text-secondary)', lineHeight: 1.5 }}>
                                说明: <span style={{ color: 'var(--text-primary)' }}>{item.description}</span>
                              </div>
                            ) : null}
                          </div>
                        </div>

                        <div
                          style={{
                            minWidth: 120,
                            padding: '10px 12px',
                            borderRadius: 12,
                            background: 'rgba(0,0,0,0.025)',
                            display: 'grid',
                            gap: 6,
                          }}
                        >
                          <div style={{ fontSize: 11, color: 'var(--text-secondary)' }}>排序</div>
                          <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--text-primary)' }}>#{item.sort}</div>
                          <div style={{ fontSize: 11, color: 'var(--text-secondary)' }}>最近更新时间</div>
                          <div style={{ fontSize: 12, color: 'var(--text-primary)', lineHeight: 1.5 }}>
                            {formatDisplayTime(item.lastFetchTime)}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </DataState>
            </div>

            {!drawerLoading && drawerTotal > 0 ? (
              <div
                style={{
                  padding: '12px 16px',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  gap: 12,
                  flexWrap: 'wrap',
                  borderTop: '1px solid var(--border-color)',
                  background: 'rgba(255,255,255,0.38)',
                }}
              >
                <Text type="secondary" style={{ fontSize: 12 }}>
                  第 {drawerPage} 页，每页 {drawerLimit} 条
                </Text>
                <Pagination
                  sizeCanChange={false}
                  current={drawerPage}
                  pageSize={drawerLimit}
                  total={drawerTotal}
                  onChange={(nextPage) => void loadSubscribeOutbounds(drawerSubscribeName, nextPage)}
                />
              </div>
            ) : null}
          </div>
        </div>
      </Drawer>
    </>
  );
}
