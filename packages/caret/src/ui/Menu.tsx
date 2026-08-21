import React, { useEffect, useRef, useState } from 'react';
import { Check } from 'lucide-react';
import { caretTokens, caretRadius } from './tokens';

export interface MenuItem {
  id: string;
  label: string;
  onSelect: () => void;
  description?: string;
  icon?: React.ReactNode;
  destructive?: boolean;
  disabled?: boolean;
  selected?: boolean;
}

export interface MenuProps {
  trigger: React.ReactElement;
  items: MenuItem[];
  align?: 'start' | 'end';
  /** "auto" flips up when the trigger is near the bottom of the viewport. */
  side?: 'auto' | 'top' | 'bottom';
  minWidth?: number;
}

export const Menu: React.FC<MenuProps> = ({
  trigger,
  items,
  align = 'start',
  side = 'auto',
  minWidth = 220,
}) => {
  const [open, setOpen] = useState(false);
  const [resolvedSide, setResolvedSide] = useState<'top' | 'bottom'>('bottom');
  const rootRef = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    if (!open) return;
    const onClick = (e: MouseEvent): void => {
      const root = rootRef.current;
      if (!root) return;
      const target = e.target;
      if (target instanceof Node && root.contains(target)) return;
      setOpen(false);
    };
    const onKey = (e: KeyboardEvent): void => {
      if (e.key !== 'Escape') return;
      e.preventDefault();
      e.stopPropagation();
      setOpen(false);
      rootRef.current?.querySelector<HTMLButtonElement>('button')?.focus();
    };
    document.addEventListener('mousedown', onClick);
    document.addEventListener('keydown', onKey);
    return (): void => {
      document.removeEventListener('mousedown', onClick);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  useEffect(() => {
    if (!open) return;
    if (side !== 'auto') {
      setResolvedSide(side);
      return;
    }
    const root = rootRef.current;
    if (!root) return;
    const rect = root.getBoundingClientRect();
    const spaceBelow = window.innerHeight - rect.bottom;
    // Menu height grows with item count; estimate ~36px per row + padding.
    const estimated = Math.min(items.length, 8) * 36 + 16;
    setResolvedSide(spaceBelow < estimated + 24 ? 'top' : 'bottom');
  }, [open, side, items.length]);

  return (
    <span ref={rootRef} style={{ position: 'relative', display: 'inline-flex' }}>
      {React.cloneElement(
        trigger as React.ReactElement<{
          onClick?: (e: React.MouseEvent) => void;
          'aria-expanded'?: boolean;
        }>,
        {
          'aria-expanded': open,
          onClick: (e: React.MouseEvent) => {
            e.stopPropagation();
            setOpen((v) => !v);
            const props = trigger.props as {
              onClick?: (e: React.MouseEvent) => void;
            };
            if (typeof props.onClick === 'function') props.onClick(e);
          },
        }
      )}
      {open && (
        <div
          role="menu"
          style={{
            position: 'absolute',
            [resolvedSide === 'top' ? 'bottom' : 'top']: 'calc(100% + 4px)',
            [align === 'end' ? 'right' : 'left']: 0,
            minWidth,
            maxWidth: 'calc(100cqw - 16px)',
            maxHeight: 'min(280px, 60vh)',
            overflowY: 'auto',
            background: caretTokens.surface,
            border: `1px solid ${caretTokens.border}`,
            borderRadius: caretRadius.lg,
            padding: 4,
            zIndex: 200,
            boxShadow: caretTokens.shadow,
          }}
        >
          {items.map((it) => (
            <button
              key={it.id}
              role={it.selected === undefined ? 'menuitem' : 'menuitemradio'}
              aria-checked={it.selected === undefined ? undefined : it.selected}
              disabled={it.disabled}
              onClick={() => {
                if (it.disabled) return;
                it.onSelect();
                setOpen(false);
              }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                minHeight: 32,
                padding: '5px 8px',
                borderRadius: caretRadius.sm,
                width: '100%',
                background: it.selected ? caretTokens.surfaceHigh : 'transparent',
                border: `1px solid ${it.selected ? caretTokens.borderStrong : 'transparent'}`,
                color: it.destructive ? caretTokens.danger : caretTokens.text,
                fontSize: 12.5,
                cursor: it.disabled ? 'not-allowed' : 'pointer',
                opacity: it.disabled ? 0.5 : 1,
                textAlign: 'left',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = caretTokens.surfaceHigh;
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = it.selected
                  ? caretTokens.surfaceHigh
                  : 'transparent';
              }}
            >
              {it.icon && <span style={{ display: 'inline-flex' }}>{it.icon}</span>}
              <span style={{ flex: 1 }}>
                <span style={{ display: 'block' }}>{it.label}</span>
                {it.description && (
                  <span
                    style={{
                      display: 'block',
                      fontSize: 11,
                      color: caretTokens.textMuted,
                      marginTop: 2,
                    }}
                  >
                    {it.description}
                  </span>
                )}
              </span>
              {it.selected && <Check aria-hidden="true" size={12} color={caretTokens.accent} />}
            </button>
          ))}
        </div>
      )}
    </span>
  );
};

export default Menu;
