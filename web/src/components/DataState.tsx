import { Button } from '@arco-design/web-react';
import type { ReactNode } from 'react';

interface Props {
  loading: boolean;
  isEmpty: boolean;
  emptyTitle: string;
  emptyDescription: string;
  createLabel?: string;
  onCreate?: () => void;
  skeletonCount?: number;
  children: ReactNode;
}

function LoadingCardGrid({ count = 6 }: { count?: number }) {
  return (
    <div className="card-grid card-grid-loading">
      {Array.from({ length: count }).map((_, index) => (
        <div key={index} className="glass-card loading-card">
          <div className="loading-line loading-line-title" />
          <div className="loading-line loading-line-tag" />
          <div className="loading-line" />
          <div className="loading-line loading-line-short" />
        </div>
      ))}
    </div>
  );
}

export default function DataState({
  loading,
  isEmpty,
  emptyTitle,
  emptyDescription,
  createLabel,
  onCreate,
  skeletonCount = 6,
  children,
}: Props) {
  if (loading) {
    return <LoadingCardGrid count={skeletonCount} />;
  }

  if (isEmpty) {
    return (
      <div className="empty-state">
        <div className="empty-state-title">{emptyTitle}</div>
        <div className="empty-state-description">{emptyDescription}</div>
        {onCreate && createLabel && (
          <Button type="primary" className="empty-state-action" onClick={onCreate}>
            {createLabel}
          </Button>
        )}
      </div>
    );
  }

  return <>{children}</>;
}
