import React from 'react';
import { LucideIcon } from 'lucide-react';
import { Button } from './Button';
import { cn } from '../lib/cn';

interface EmptyStateAction {
  label: string;
  onClick: () => void;
}

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  action: EmptyStateAction;
  className?: string;
}

const EmptyStateIllustration = ({ icon: Icon }: { icon: LucideIcon }): React.JSX.Element => (
  <div className="relative mb-4 h-[120px] w-[120px] shrink-0">
    <svg
      width="120"
      height="120"
      viewBox="0 0 120 120"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <rect x="20" y="14" width="80" height="98" rx="10" className="fill-accent/10" />
      <rect x="36" y="8" width="48" height="16" rx="6" className="fill-accent/5" />
      <rect x="34" y="40" width="52" height="6" rx="3" className="fill-accent/20" />
      <rect x="34" y="56" width="38" height="6" rx="3" className="fill-accent/20" />
    </svg>
    <div className="absolute inset-x-0 bottom-2 flex items-center justify-center">
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-accent/20">
        <Icon size={22} className="text-accent" aria-hidden="true" />
      </div>
    </div>
  </div>
);

// Every empty state gets an illustration and a next step, never a dead end -
// see the "no empty states without a call to action" product rule.
const EmptyState = ({
  icon,
  title,
  description,
  action,
  className,
}: EmptyStateProps): React.JSX.Element => (
  <div
    className={cn(
      'flex flex-col items-center justify-center px-6 py-16 text-center',
      className
    )}
  >
    <EmptyStateIllustration icon={icon} />
    <h3 className="mb-1 text-sm font-semibold">{title}</h3>
    {description && <p className="mb-4 max-w-sm text-xs opacity-60">{description}</p>}
    <Button variant="accent" size="sm" onClick={action.onClick}>
      {action.label}
    </Button>
  </div>
);

export { EmptyState };
export type { EmptyStateProps, EmptyStateAction };
