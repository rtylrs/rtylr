import React, { useCallback, useEffect, useId, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
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

interface TooltipPosition {
  left: number;
  top: number;
  side: TooltipSide;
}

const viewportPadding = 8;
const tooltipGap = 10;

const arrowClasses: Record<TooltipSide, string> = {
  top: 'left-1/2 top-full -translate-x-1/2 border-l-transparent border-r-transparent border-b-transparent',
  right:
    'right-full top-1/2 -translate-y-1/2 border-l-transparent border-t-transparent border-b-transparent',
  bottom:
    'bottom-full left-1/2 -translate-x-1/2 border-l-transparent border-r-transparent border-t-transparent',
  left: 'left-full top-1/2 -translate-y-1/2 border-r-transparent border-t-transparent border-b-transparent',
};

const arrowStyles: Record<TooltipSide, React.CSSProperties> = {
  top: { borderTopColor: 'var(--dash-text, rgb(9, 9, 11))' },
  right: { borderRightColor: 'var(--dash-text, rgb(9, 9, 11))' },
  bottom: { borderBottomColor: 'var(--dash-text, rgb(9, 9, 11))' },
  left: { borderLeftColor: 'var(--dash-text, rgb(9, 9, 11))' },
};

const getOppositeSide = (side: TooltipSide): TooltipSide => {
  switch (side) {
    case 'top':
      return 'bottom';
    case 'right':
      return 'left';
    case 'bottom':
      return 'top';
    case 'left':
      return 'right';
  }
};

const getPositionForSide = (
  triggerRect: DOMRect,
  tooltipRect: DOMRect,
  side: TooltipSide
): Omit<TooltipPosition, 'side'> => {
  switch (side) {
    case 'top':
      return {
        left: triggerRect.left + triggerRect.width / 2 - tooltipRect.width / 2,
        top: triggerRect.top - tooltipRect.height - tooltipGap,
      };
    case 'right':
      return {
        left: triggerRect.right + tooltipGap,
        top: triggerRect.top + triggerRect.height / 2 - tooltipRect.height / 2,
      };
    case 'bottom':
      return {
        left: triggerRect.left + triggerRect.width / 2 - tooltipRect.width / 2,
        top: triggerRect.bottom + tooltipGap,
      };
    case 'left':
      return {
        left: triggerRect.left - tooltipRect.width - tooltipGap,
        top: triggerRect.top + triggerRect.height / 2 - tooltipRect.height / 2,
      };
  }
};

const fitsInViewport = (position: Omit<TooltipPosition, 'side'>, tooltipRect: DOMRect): boolean =>
  position.left >= viewportPadding &&
  position.top >= viewportPadding &&
  position.left + tooltipRect.width <= window.innerWidth - viewportPadding &&
  position.top + tooltipRect.height <= window.innerHeight - viewportPadding;

const clampPosition = (
  position: Omit<TooltipPosition, 'side'>,
  tooltipRect: DOMRect
): Omit<TooltipPosition, 'side'> => ({
  left: Math.min(
    Math.max(position.left, viewportPadding),
    Math.max(viewportPadding, window.innerWidth - tooltipRect.width - viewportPadding)
  ),
  top: Math.min(
    Math.max(position.top, viewportPadding),
    Math.max(viewportPadding, window.innerHeight - tooltipRect.height - viewportPadding)
  ),
});

const Tooltip: React.FC<TooltipProps> = ({
  children,
  content,
  side = 'top',
  disabled = false,
  className,
  contentClassName,
}): React.JSX.Element => {
  const tooltipId = useId();
  const triggerRef = useRef<HTMLSpanElement>(null);
  const tooltipRef = useRef<HTMLSpanElement>(null);
  const [open, setOpen] = useState(false);
  const [position, setPosition] = useState<TooltipPosition | null>(null);
  const isVisible = open && !disabled && Boolean(content);

  const updatePosition = useCallback(() => {
    if (!triggerRef.current || !tooltipRef.current) {
      return;
    }

    const triggerRect = triggerRef.current.getBoundingClientRect();
    const tooltipRect = tooltipRef.current.getBoundingClientRect();
    const preferredPosition = getPositionForSide(triggerRect, tooltipRect, side);
    const oppositeSide = getOppositeSide(side);
    const oppositePosition = getPositionForSide(triggerRect, tooltipRect, oppositeSide);

    if (fitsInViewport(preferredPosition, tooltipRect)) {
      setPosition({ ...preferredPosition, side });
      return;
    }

    if (fitsInViewport(oppositePosition, tooltipRect)) {
      setPosition({ ...oppositePosition, side: oppositeSide });
      return;
    }

    setPosition({ ...clampPosition(preferredPosition, tooltipRect), side });
  }, [side]);

  useEffect(() => {
    if (!isVisible) {
      setPosition(null);
      return undefined;
    }

    updatePosition();
    window.addEventListener('resize', updatePosition);
    window.addEventListener('scroll', updatePosition, true);

    return () => {
      window.removeEventListener('resize', updatePosition);
      window.removeEventListener('scroll', updatePosition, true);
    };
  }, [isVisible, updatePosition]);

  if (disabled || !content) {
    return <>{children}</>;
  }

  return (
    <span
      ref={triggerRef}
      className={cn('relative inline-flex', className)}
      aria-describedby={isVisible ? tooltipId : undefined}
      tabIndex={0}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
      onFocus={() => setOpen(true)}
      onBlur={() => setOpen(false)}
    >
      {children}
      {isVisible && typeof document !== 'undefined'
        ? createPortal(
            <span
              ref={tooltipRef}
              id={tooltipId}
              role="tooltip"
              className={cn(
                'pointer-events-none fixed z-[2147483647] w-max min-w-44 max-w-72 rounded-md px-2.5 py-1.5 text-left text-xs font-medium leading-snug shadow-lg transition-opacity duration-150',
                position ? 'opacity-100' : 'opacity-0',
                contentClassName
              )}
              style={{
                left: position?.left ?? -9999,
                top: position?.top ?? -9999,
                background: 'var(--dash-text, rgb(9, 9, 11))',
                color: 'var(--dash-bg, #fff)',
                border: '1px solid var(--dash-border, rgba(0, 0, 0, 0.08))',
              }}
            >
              {content}
              <span
                className={cn('absolute h-0 w-0 border-4', arrowClasses[position?.side ?? side])}
                style={arrowStyles[position?.side ?? side]}
              />
            </span>,
            document.body
          )
        : null}
    </span>
  );
};

export { Tooltip };
