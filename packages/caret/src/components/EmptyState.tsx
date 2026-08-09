import React from 'react';

export interface EmptyStateProps {
  label?: string;
}

/** A deliberately quiet starting canvas. The composer carries the visible prompt. */
export const EmptyState: React.FC<EmptyStateProps> = ({
  label = 'No messages yet. Ask Caret a question about this page.',
}) => (
  <div className="caret-empty-state" style={{ flex: 1, minHeight: 0 }}>
    <span className="caret-visually-hidden">{label}</span>
  </div>
);

export default EmptyState;
