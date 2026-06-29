import React, { useId, useState } from 'react';
import { cn } from '../../lib/utils';

export type TooltipSide = 'top' | 'right' | 'bottom' | 'left';

export interface TooltipProps {
  children: React.ReactNode;
  content: React.ReactNode;
  side?: TooltipSide;
  disabled?: boolean;
  className?: string;
  contentClassName?: string;
}

const sideClasses: Record<TooltipSide, string> = {
  top: 'bottom-full left-1/2 mb-2 -translate-x-1/2',
  right: 'left-full top-1/2 ml-2 -translate-y-1/2',
  bottom: 'left-1/2 top-full mt-2 -translate-x-1/2',
  left: 'right-full top-1/2 mr-2 -translate-y-1/2',
};

const arrowClasses: Record<TooltipSide, string> = {
  top: 'left-1/2 top-full -translate-x-1/2 border-l-transparent border-r-transparent border-b-transparent',
  right:
    'right-full top-1/2 -translate-y-1/2 border-l-transparent border-t-transparent border-b-transparent',
  bottom:
    'bottom-full left-1/2 -translate-x-1/2 border-l-transparent border-r-transparent border-t-transparent',
  left: 'left-full top-1/2 -translate-y-1/2 border-r-transparent border-t-transparent border-b-transparent',
};

const arrowStyles: Record<TooltipSide, React.CSSProperties> = {
  top: { borderTopColor: 'var(--dash-text)' },
  right: { borderRightColor: 'var(--dash-text)' },
  bottom: { borderBottomColor: 'var(--dash-text)' },
  left: { borderLeftColor: 'var(--dash-text)' },
};

const Tooltip: React.FC<TooltipProps> = ({
  children,
  content,
  side = 'top',
  disabled = false,
  className,
  contentClassName,
}): React.JSX.Element => {
  const tooltipId = useId();
  const [open, setOpen] = useState(false);
  const isVisible = open && !disabled && Boolean(content);

  if (disabled || !content) {
    return <>{children}</>;
  }

  return (
    <span
      className={cn('relative inline-flex', className)}
      aria-describedby={isVisible ? tooltipId : undefined}
      tabIndex={0}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
      onFocus={() => setOpen(true)}
      onBlur={() => setOpen(false)}
    >
      {children}
      <span
        id={tooltipId}
        role="tooltip"
        className={cn(
          'pointer-events-none absolute z-[100] w-max max-w-64 rounded-md px-2.5 py-1.5 text-xs font-medium leading-snug shadow-lg transition-opacity duration-150',
          sideClasses[side],
          isVisible ? 'opacity-100' : 'opacity-0',
          contentClassName
        )}
        style={{
          background: 'var(--dash-text)',
          color: 'var(--dash-bg)',
          border: '1px solid var(--dash-border)',
        }}
      >
        {content}
        <span
          className={cn('absolute h-0 w-0 border-4', arrowClasses[side])}
          style={arrowStyles[side]}
        />
      </span>
    </span>
  );
};

export { Tooltip };
