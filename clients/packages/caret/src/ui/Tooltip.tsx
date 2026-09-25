import React, { useCallback, useId, useLayoutEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { caretTokens, caretRadius } from './tokens';

type TooltipSide = 'top' | 'bottom';

export interface TooltipProps {
  label: string;
  side?: TooltipSide;
  children: React.ReactElement;
}

interface TooltipPosition {
  top: number;
  left: number;
}

const tooltipGap = 8;
const viewportPadding = 8;

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

export const Tooltip: React.FC<TooltipProps> = ({ label, side = 'top', children }) => {
  const [show, setShow] = useState(false);
  const [position, setPosition] = useState<TooltipPosition | null>(null);
  const triggerRef = useRef<HTMLSpanElement>(null);
  const tooltipRef = useRef<HTMLSpanElement>(null);
  const tooltipId = useId();

  const updatePosition = useCallback(() => {
    if (!triggerRef.current || !tooltipRef.current) return;

    const triggerRect = triggerRef.current.getBoundingClientRect();
    const tooltipRect = tooltipRef.current.getBoundingClientRect();
    const preferredTop =
      side === 'top'
        ? triggerRect.top - tooltipRect.height - tooltipGap
        : triggerRect.bottom + tooltipGap;
    const oppositeTop =
      side === 'top'
        ? triggerRect.bottom + tooltipGap
        : triggerRect.top - tooltipRect.height - tooltipGap;
    const fitsPreferred =
      preferredTop >= viewportPadding &&
      preferredTop + tooltipRect.height <= window.innerHeight - viewportPadding;
    const rawTop = fitsPreferred ? preferredTop : oppositeTop;
    const top = clamp(
      rawTop,
      viewportPadding,
      window.innerHeight - tooltipRect.height - viewportPadding
    );
    const centeredLeft = triggerRect.left + triggerRect.width / 2 - tooltipRect.width / 2;
    const left = clamp(
      centeredLeft,
      viewportPadding,
      window.innerWidth - tooltipRect.width - viewportPadding
    );

    setPosition({ top, left });
  }, [side]);

  useLayoutEffect(() => {
    if (!show) return;
    updatePosition();
    window.addEventListener('resize', updatePosition);
    window.addEventListener('scroll', updatePosition, true);
    return () => {
      window.removeEventListener('resize', updatePosition);
      window.removeEventListener('scroll', updatePosition, true);
    };
  }, [show, updatePosition]);

  return (
    <span
      ref={triggerRef}
      aria-describedby={show ? tooltipId : undefined}
      style={{ display: 'inline-flex' }}
      onMouseEnter={() => setShow(true)}
      onMouseLeave={() => setShow(false)}
      onFocus={() => setShow(true)}
      onBlur={() => setShow(false)}
    >
      {children}
      {show &&
        typeof document !== 'undefined' &&
        createPortal(
          <span
            ref={tooltipRef}
            id={tooltipId}
            role="tooltip"
            style={{
              position: 'fixed',
              top: position?.top ?? -9999,
              left: position?.left ?? -9999,
              background: caretTokens.surfaceHigh,
              color: caretTokens.text,
              fontSize: 11,
              padding: '4px 6px',
              borderRadius: caretRadius.sm,
              border: `1px solid ${caretTokens.border}`,
              whiteSpace: 'nowrap',
              pointerEvents: 'none',
              zIndex: 2147483647,
              opacity: position ? 1 : 0,
            }}
          >
            {label}
          </span>,
          document.body
        )}
    </span>
  );
};

export default Tooltip;
