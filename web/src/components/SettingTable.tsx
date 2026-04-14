import { Button, Grid, Typography } from '@arco-design/web-react';
import type { Setting } from '../types';

const { Row, Col } = Grid;
const { Title, Text } = Typography;

interface Props {
  data: Setting[];
  deletingKey?: string | null;
  onEdit: (record: Setting) => void;
  onDelete: (record: Setting) => void;
}

export default function SettingTable({ data, deletingKey, onEdit, onDelete }: Props) {
  return (
    <div style={{ marginBottom: 20 }}>
      <Row gutter={[16, 16]}>
        {data.map((item) => (
          <Col key={item.key} xs={24} sm={12} lg={6}>
            <div className="glass-card">
              <div className="card-header" style={{ padding: '12px 16px 8px' }}>
                <div style={{ flex: 1 }}>
                  <Title heading={6} style={{ margin: 0, fontSize: 14 }}>{item.key}</Title>
                </div>
              </div>
              <div className="card-content" style={{ padding: '0 16px 12px' }}>
                <Text type="secondary" style={{ fontSize: 11 }}>值</Text>
                <div style={{ 
                  fontSize: 12, 
                  marginTop: 4, 
                  wordBreak: 'break-all',
                  fontFamily: 'monospace',
                  background: 'rgba(0,0,0,0.025)',
                  padding: '6px 10px',
                  borderRadius: 6,
                  maxHeight: '60px',
                  overflow: 'auto'
                }}>
                  {item.value}
                </div>
              </div>
              <div className="card-footer" style={{ padding: '8px 16px' }}>
                <div className="card-action-row">
                  <Button type="text" size="small" style={{ fontSize: 12, padding: '0 4px' }} onClick={() => onEdit(item)}>编辑</Button>
                  <Button
                    type="text"
                    size="small"
                    status="danger"
                    loading={deletingKey === item.key}
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
