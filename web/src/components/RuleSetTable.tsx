import { Button, Grid, Typography, Tag, Space, Select } from '@arco-design/web-react';
import type { RuleSet, NodeGroup } from '../types';

const { Row, Col } = Grid;
const { Title, Text } = Typography;

interface Props {
  data: RuleSet[];
  nodeGroups: NodeGroup[];
  deletingKey?: string | null;
  togglingKey?: string | null;
  onEdit: (record: RuleSet) => void;
  onDelete: (record: RuleSet) => void;
  onChangeOutbound: (record: RuleSet, newOutbound: string) => void;
}

export default function RuleSetTable({ data, nodeGroups, deletingKey, togglingKey, onEdit, onDelete, onChangeOutbound }: Props) {
  return (
    <div style={{ marginBottom: 20 }}>
      <Row gutter={[16, 16]}>
        {data.map((item) => (
          <Col key={item.tag} xs={24} sm={12} lg={6}>
            <div className="glass-card">
              <div className="card-header" style={{ padding: '12px 16px 8px' }}>
                <div style={{ flex: 1 }}>
                  <Title heading={6} style={{ margin: 0, fontSize: 15 }}>{item.name}</Title>
                  <Space style={{ marginTop: 4 }}>
                    <Tag size="small" color="purple" style={{ scale: '0.85', transformOrigin: 'left center' }}>{item.ruleSetType}</Tag>
                    <Tag size="small" color="cyan" style={{ scale: '0.85', transformOrigin: 'left center' }}>{item.format}</Tag>
                  </Space>
                </div>
                <div style={{ fontSize: 11, fontWeight: 700, opacity: 0.4 }}>
                  #{item.sort}
                </div>
              </div>
              <div className="card-content" style={{ padding: '0 16px 12px' }}>
                <div style={{ marginBottom: 8 }}>
                  <Text type="secondary" style={{ fontSize: 11 }}>Tag</Text>
                  <div style={{ fontSize: 12, marginTop: 2, fontWeight: 500 }}>{item.tag}</div>
                </div>
                <div style={{ marginBottom: 8 }}>
                  <Text type="secondary" style={{ fontSize: 11 }}>默认出站</Text>
                  <Select
                    size="small"
                    value={item.outbound || ''}
                    loading={togglingKey === item.tag}
                    onChange={(value) => onChangeOutbound(item, value)}
                    style={{ width: '100%', marginTop: 4 }}
                    placeholder="选择出站"
                    allowClear
                  >
                    {nodeGroups.map((ng) => (
                      <Select.Option key={ng.tag} value={ng.tag}>
                        {ng.name} ({ng.tag})
                      </Select.Option>
                    ))}
                  </Select>
                </div>
                {item.url && (
                  <div style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    <Text type="secondary" style={{ fontSize: 11 }}>远程地址</Text>
                    <div style={{ fontSize: 11, marginTop: 2, opacity: 0.7 }}>
                      {item.url}
                    </div>
                  </div>
                )}
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
