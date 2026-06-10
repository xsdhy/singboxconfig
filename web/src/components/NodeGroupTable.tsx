import { Button, Grid, Typography, Tag, Space, Select } from '@arco-design/web-react';
import type { NodeGroup } from '../types';

const { Row, Col } = Grid;
const { Title, Text } = Typography;

// deviceOverrideItem 表示一条解析后的设备级类型覆盖规则，用于列表卡片展示。
interface deviceOverrideItem {
  deviceCode: string; // 设备编码。
  groupType: string; // 目标分组类型（selector / urltest）。
}

// parseOverrides 把后端的规则字符串解析为展示用数组，解析规则与后端 ParseDeviceTypeOverrides 对齐：
// 逗号分隔多条规则，第一个冒号分隔设备编码与类型，忽略空白与不合法项。
function parseOverrides(raw?: string): deviceOverrideItem[] {
  if (!raw || !raw.trim()) return [];
  const items: deviceOverrideItem[] = [];
  for (const rule of raw.split(',')) {
    const idx = rule.indexOf(':');
    if (idx < 0) continue;
    const deviceCode = rule.slice(0, idx).trim();
    const groupType = rule.slice(idx + 1).trim();
    if (!deviceCode || !groupType) continue;
    items.push({ deviceCode, groupType });
  }
  return items;
}

interface Props {
  data: NodeGroup[];
  deletingKey?: string | null;
  togglingKey?: string | null;
  onEdit: (record: NodeGroup) => void;
  onDelete: (record: NodeGroup) => void;
  onToggleGroupType: (record: NodeGroup, newType: string) => void;
}

export default function NodeGroupTable({ data, deletingKey, togglingKey, onEdit, onDelete, onToggleGroupType }: Props) {
  return (
    <div style={{ marginBottom: 20 }}>
      <Row gutter={[16, 16]}>
        {data.map((item) => {
          const overrides = parseOverrides(item.deviceTypeOverrides);
          return (
          <Col key={item.tag} xs={24} sm={12} lg={6}>
            <div className="glass-card">
              <div className="card-header" style={{ padding: '12px 16px 8px' }}>
                <div style={{ flex: 1 }}>
                  <Title heading={6} style={{ margin: 0, fontSize: 15 }}>{item.name}</Title>
                  <div style={{ marginTop: 6, display: 'flex', alignItems: 'center', gap: 8 }}>
                    <Text type="secondary" style={{ fontSize: 11 }}>分组类型</Text>
                    <Select
                      size="small"
                      value={item.groupType}
                      loading={togglingKey === item.tag}
                      onChange={(value) => onToggleGroupType(item, value)}
                      style={{ width: 100 }}
                    >
                      <Select.Option value="selector">selector</Select.Option>
                      <Select.Option value="urltest">urltest</Select.Option>
                    </Select>
                    <Tag size="small" style={{ scale: '0.85' }}>{item.tag}</Tag>
                  </div>
                </div>
              </div>
              <div className="card-content" style={{ padding: '0 16px 12px' }}>
                <Space direction="vertical" size={4} style={{ width: '100%' }}>
                  {item.include && (
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>包含</Text>
                      <div style={{ fontSize: 12, marginTop: 2, color: 'var(--text-primary)', opacity: 0.8, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {item.include}
                      </div>
                    </div>
                  )}
                  {item.exclude && (
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>排除</Text>
                      <div style={{ fontSize: 12, marginTop: 2, color: 'var(--text-primary)', opacity: 0.8, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {item.exclude}
                      </div>
                    </div>
                  )}
                  {overrides.length > 0 && (
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>设备级类型覆盖</Text>
                      <div style={{ marginTop: 4, display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                        {overrides.map((o) => (
                          <Tag
                            key={o.deviceCode}
                            size="small"
                            color={o.groupType === 'urltest' ? 'arcoblue' : 'green'}
                            style={{ fontSize: 11 }}
                          >
                            {o.deviceCode} → {o.groupType}
                          </Tag>
                        ))}
                      </div>
                    </div>
                  )}
                </Space>
              </div>
              <div className="card-footer" style={{ padding: '8px 16px' }}>
                <div className="card-action-row">
                  <Button type="text" size="small" style={{ fontSize: 12, padding: '0 4px' }} onClick={() => onEdit(item)}>编辑</Button>
                  <Button
                    type="text"
                    size="small"
                    status="danger"
                    loading={deletingKey === item.tag}
                    style={{ fontSize: 12, padding: '0 4px' }}
                    onClick={() => onDelete(item)}
                  >
                    删除
                  </Button>
                </div>
              </div>
            </div>
          </Col>
          );
        })}
      </Row>
    </div>
  );
}
