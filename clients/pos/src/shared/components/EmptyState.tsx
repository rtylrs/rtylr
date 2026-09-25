import React from 'react';
import { Inbox } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  message?: string;
  action?: React.ReactNode;
}

const EmptyState: React.FC<EmptyStateProps> = ({ icon: Icon = Inbox, title, message, action }) => {
  return (
    <div className="flex flex-col items-center justify-center h-full min-h-[16rem] gap-3 text-pos-muted text-center px-6">
      <div className="flex h-14 w-14 items-center justify-center rounded-xl border border-pos-border bg-white/5 text-pos-muted">
        <Icon className="h-7 w-7" aria-hidden="true" strokeWidth={1.8} />
      </div>
      <h3 className="text-base font-semibold text-pos-text">{title}</h3>
      {message && <p className="text-sm max-w-md">{message}</p>}
      {action && <div className="mt-2">{action}</div>}
    </div>
  );
};

export default EmptyState;
