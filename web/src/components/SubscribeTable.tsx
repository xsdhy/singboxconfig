import { Button, Tag, Grid, Typography, Switch } from '@arco-design/web-react';
import type { Subscribe } from '../types';
import { summarizeSubscribeCache } from '../utils/subscribeOutbound';

const { Row, Col } = Grid;
const { Title, Text } = Typography;

interface Props {
  data: Subscribe[];
  deletingKey?: string | null;
  refreshingKey?: string | null;
  viewingKey?: string | null;
  togglingKey?: string | null;
  onEdit: (record: Subscribe) => void;
  onDelete: (record: Subscribe) => void;
  onRefreshOutbounds: (record: Subscribe) => void;
  onViewOutbounds: (record: Subscribe) => void;
  onToggleStatus: (record: Subscribe, newStatus: boolean) => void;
}

export default function SubscribeTable({
  data,
  deletingKey,
  refreshingKey,
  viewingKey,
  togglingKey,
  onEdit,
  onDelete,
  onRefreshOutbounds,
  onViewOutbounds,
  onToggleStatus,
}: Props) {
  return (
    <div style={{ marginBottom: 20 }}>
      <Row gutter={[16, 16]}>
        {data.map((item) => (
          <Col key={item.name} xs={24} sm={12} lg={8} xl={6}>
            <div className="glass-card">
              <div className="card-header" style={{ padding: '12px 16px 8px' }}>
                <div style={{ flex: 1 }}>
                  <Title heading={6} style={{ margin: 0, fontSize: 15 }}>{item.name}</Title>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 6 }}>
                    <Switch
                      size="small"
                      checked={item.status}
                      loading={togglingKey === item.name}
                      onChange={(checked) => onToggleStatus(item, checked)}
                    />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {item.status ? '已启用' : '已禁用'}
                    </Text>
                  </div>
                </div>
              </div>
              <div className="card-content" style={{ padding: '0 16px 12px' }}>
                <div style={{ display: 'grid', gap: 10 }}>
                  <div>
                    <Text type="secondary" style={{ fontSize: 11 }}>订阅地址</Text>
                    <div style={{ wordBreak: 'break-all', fontSize: 12, marginTop: 2, color: 'var(--text-primary)', opacity: 0.8 }}>
                      {item.url}
                    </div>
                  </div>
                  <div>
                    <Text type="secondary" style={{ fontSize: 11 }}>可见设备</Text>
                    <div style={{ fontSize: 12 }}>{item.visibleDevices || '全部设备'}</div>
                  </div>
                  <div style={{ background: 'rgba(0,0,0,0.025)', borderRadius: 10, padding: '8px 10px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8, marginBottom: 4 }}>
                      <Text type="secondary" style={{ fontSize: 11 }}>Outbound 缓存</Text>
                      <Tag color={summarizeSubscribeCache(item).color} size="small">{summarizeSubscribeCache(item).label}</Tag>
                    </div>
                    <div style={{ fontSize: 12, lineHeight: 1.5 }}>
                      {summarizeSubscribeCache(item).detail}
                    </div>
                    <div style={{ marginTop: 6, fontSize: 11, color: 'var(--text-secondary)' }}>
                      缓存时长：{item.outboundCacheDuration ?? 0} 分钟
                    </div>
                  </div>
                </div>
              </div>
              <div className="card-footer" style={{ padding: '8px 16px' }}>
                <div className="card-action-row" style={{ flexWrap: 'wrap' }}>
                  <Button
                    type="text"
                    size="small"
                    loading={refreshingKey === item.name}
                    style={{ fontSize: 12, padding: '0 4px' }}
                    onClick={() => onRefreshOutbounds(item)}
                  >
                    立即刷新
                  </Button>
                  <Button
                    type="text"
                    size="small"
                    loading={viewingKey === item.name}
                    style={{ fontSize: 12, padding: '0 4px' }}
                    onClick={() => onViewOutbounds(item)}
                  >
                    查看节点
                  </Button>
                  <Button type="text" size="small" style={{ fontSize: 12, padding: '0 4px' }} onClick={() => onEdit(item)}>编辑</Button>
                  <Button
                    type="text"
                    size="small"
                    status="danger"
                    loading={deletingKey === item.name}
                    style={{ fontSize: 12, padding: '0 4px' }}
                    onClick={() => onDelete(item)}
                  >
                    删除
                  </Button>
                </div>
              </div>
            </div>
          </Col>
        ))}
      </Row>
    </div>
  );
}
