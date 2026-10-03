import React from 'react';
import type { LucideIcon } from 'lucide-react';
import { Button, type ButtonProps } from './Button';
import { cn } from '../lib/cn';

interface EmptyStateAction {
  label: string;
  onClick: () => void;
  icon?: LucideIcon;
  variant?: ButtonProps['variant'];
  size?: ButtonProps['size'];
}

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  action?: EmptyStateAction;
  className?: string;
  // 'compact' fits dashboard widgets, kanban columns and other tight
  // spaces that still need an illustration and a CTA, just smaller.
  size?: 'default' | 'compact';
}

const ILLUSTRATION_PX: Record<'default' | 'compact', number> = {
  default: 120,
  compact: 64,
};

const EmptyStateIllustration = ({
  icon: Icon,
  size,
}: {
  icon: LucideIcon;
  size: 'default' | 'compact';
}): React.JSX.Element => {
  const px = ILLUSTRATION_PX[size];
  const iconWrapperPx = size === 'compact' ? 28 : 48;
  const iconSize = size === 'compact' ? 14 : 22;
  return (
    <div className="relative mb-3 shrink-0" style={{ height: px, width: px }}>
      <svg
        width={px}
        height={px}
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
        <div
          className="flex items-center justify-center rounded-full bg-accent/20"
          style={{ height: iconWrapperPx, width: iconWrapperPx }}
        >
          <Icon size={iconSize} className="text-accent" aria-hidden="true" />
        </div>
      </div>
    </div>
  );
};

const EmptyState = ({
  icon,
  title,
  description,
  action,
  className,
  size = 'default',
}: EmptyStateProps): React.JSX.Element => (
  <div
    className={cn(
      'flex flex-col items-center justify-center text-center',
      size === 'compact' ? 'px-3 py-6' : 'px-6 py-16',
      className
    )}
  >
    <EmptyStateIllustration icon={icon} size={size} />
    <h3 className={cn('mb-1 font-semibold', size === 'compact' ? 'text-xs' : 'text-sm')}>
      {title}
    </h3>
    {description && (
      <p
        className={cn(
          'mb-3 max-w-sm opacity-60',
          size === 'compact' ? 'text-[11px]' : 'text-xs mb-4'
        )}
      >
        {description}
      </p>
    )}
    {/* Button's own xs/sm sizes are 30/32px tall - short of the 44px
       touch-target minimum. Override the height explicitly on mobile and
       let it shrink back down from md: up, where a mouse doesn't need it. */}
    {action && (
      <Button
        variant={action.variant ?? 'accent'}
        size={action.size ?? (size === 'compact' ? 'xs' : 'sm')}
        onClick={action.onClick}
        prefix={
          action.icon ? (
            <action.icon
              size={14}
              aria-hidden="true"
              className={action.variant === 'outline' ? 'text-accent' : undefined}
            />
          ) : undefined
        }
        className={
          action.size === 'md'
            ? 'h-11 md:h-10'
            : size === 'compact'
              ? 'h-11 md:h-[30px]'
              : 'h-11 md:h-8'
        }
      >
        {action.label}
      </Button>
    )}
  </div>
);

export { EmptyState };
export type { EmptyStateProps, EmptyStateAction };
