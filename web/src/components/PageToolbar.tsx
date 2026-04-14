import { Button, Space, Typography } from '@arco-design/web-react';
import type { ReactNode } from 'react';

const { Text } = Typography;

interface Props {
  summary: string;
  count?: number;
  countLabel?: string;
  primaryActionLabel?: string;
  primaryActionLoading?: boolean;
  onPrimaryAction?: () => void;
  onRefresh?: () => void;
  refreshing?: boolean;
  extraContent?: ReactNode;
}

export default function PageToolbar({
  summary,
  count,
  countLabel = '条记录',
  primaryActionLabel,
  primaryActionLoading = false,
  onPrimaryAction,
  onRefresh,
  refreshing = false,
  extraContent,
}: Props) {
  return (
    <div className="page-toolbar">
      <div className="page-toolbar-summary">
        <Text type="secondary" style={{ fontSize: 14, fontWeight: 500 }}>
          {summary}
        </Text>
        {typeof count === 'number' && (
          <div className="page-toolbar-count">
            {count} {countLabel}
          </div>
        )}
      </div>
      <Space wrap>
        {extraContent}
        {onRefresh && (
          <Button size="small" loading={refreshing} onClick={onRefresh} style={{ borderRadius: 8 }}>
            刷新
          </Button>
        )}
        {onPrimaryAction && primaryActionLabel && (
          <Button
            type="primary"
            size="small"
            loading={primaryActionLoading}
            onClick={onPrimaryAction}
            style={{ borderRadius: 8 }}
          >
            {primaryActionLabel}
          </Button>
        )}
      </Space>
    </div>
  );
}
