import React from 'react';

interface EmptyStateProps {
  icon?: string;
  title: string;
  message?: string;
  action?: React.ReactNode;
}

const EmptyState: React.FC<EmptyStateProps> = ({ icon = '📭', title, message, action }) => (
  <div className="flex flex-col items-center justify-center h-full min-h-[16rem] gap-3 text-pos-muted text-center px-6">
    <span className="text-5xl">{icon}</span>
    <h3 className="text-base font-semibold text-pos-text">{title}</h3>
    {message && <p className="text-sm max-w-md">{message}</p>}
    {action && <div className="mt-2">{action}</div>}
  </div>
);

export default EmptyState;
