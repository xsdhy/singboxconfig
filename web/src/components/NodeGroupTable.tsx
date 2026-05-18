import { Button, Grid, Typography, Tag, Space, Select } from '@arco-design/web-react';
import type { NodeGroup } from '../types';

const { Row, Col } = Grid;
const { Title, Text } = Typography;

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
        {data.map((item) => (
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
        ))}
      </Row>
    </div>
  );
}
