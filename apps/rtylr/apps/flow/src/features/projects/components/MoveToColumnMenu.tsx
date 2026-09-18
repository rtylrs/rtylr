import React, { useCallback, useEffect, useId, useLayoutEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { useTranslation } from 'react-i18next';
import { CornerUpRight, Check } from 'lucide-react';

export interface MoveTarget {
  id: string;
  name: string;
}

interface MoveToColumnMenuProps {
  taskTitle: string;
  targets: readonly MoveTarget[];
  currentColumnId: string | null;
  onMove: (columnId: string) => void;
  disabled?: boolean;
}

/**
 * The single-pointer and keyboard alternative to dragging a card (B2).
 *
 * WCAG 2.2 AA 2.5.7 requires that anything achievable by dragging is also
 * achievable without it. This is that path, and it is not a fallback: it is
 * reachable by tab, so it is the *only* way a keyboard user moves a card.
 *
 * A real <button> with aria-haspopup/aria-expanded rather than a hover-revealed
 * affordance, because the UI standards addendum prohibits actions hidden solely
 * behind hover (rule 6).
 */
const MoveToColumnMenu = ({
  taskTitle,
  targets,
  currentColumnId,
  onMove,
  disabled,
}: MoveToColumnMenuProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [position, setPosition] = useState<{ top: number; left: number } | null>(null);
  const menuId = useId();
  const triggerRef = useRef<HTMLButtonElement>(null);
  const rootRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  // Queried from the menu, not the root: the menu is portalled to the body, so
  // it is no longer a descendant of this component's own element.
  //
  // Only the enabled items take focus. The caller's current column renders
  // disabled and is usually first, so focusing "the first item" would land on
  // something unfocusable and leave focus on the body.
  const enabledItems = (): HTMLButtonElement[] =>
    Array.from(
      menuRef.current?.querySelectorAll<HTMLButtonElement>('[data-move-item]:not([disabled])') ?? []
    );

  // Close on outside press and on Escape. Escape is listened for on the root
  // rather than the document so it does not swallow the key for whatever else
  // is open behind this.
  useEffect(() => {
    if (!open) return;
    const onDocPointerDown = (e: PointerEvent): void => {
      const target = e.target as Node;
      // Both, for the same reason as enabledItems: the menu lives in a portal.
      if (rootRef.current?.contains(target) || menuRef.current?.contains(target)) return;
      setOpen(false);
    };
    document.addEventListener('pointerdown', onDocPointerDown);
    return () => document.removeEventListener('pointerdown', onDocPointerDown);
  }, [open]);

  // The menu is portalled to the body because the column's card list is an
  // overflow-y-auto scroller: rendered in place, the menu was clipped at the
  // card's edge and only the first item or two were reachable.
  //
  // Portalling means position has to be measured rather than inherited. Flip
  // above the trigger when there is not enough room below, so a card near the
  // bottom of a column does not open a menu off-screen.
  const place = useCallback((): void => {
    const trigger = triggerRef.current;
    if (!trigger) return;
    const r = trigger.getBoundingClientRect();
    const height = menuRef.current?.offsetHeight ?? 0;
    const below = window.innerHeight - r.bottom;
    const top = height > 0 && below < height + 8 ? r.top - height - 4 : r.bottom + 4;
    const width = menuRef.current?.offsetWidth ?? 160;
    // Keep it on screen horizontally; `end`-aligned to the trigger otherwise.
    const left = Math.min(Math.max(8, r.right - width), window.innerWidth - width - 8);
    setPosition({ top, left });
  }, []);

  useLayoutEffect(() => {
    if (!open) {
      setPosition(null);
      return;
    }
    place();
    // A scroll or resize while the menu is open would leave it detached from its
    // card, and the board scrolls in both axes.
    const onReflow = (): void => setOpen(false);
    window.addEventListener('scroll', onReflow, true);
    window.addEventListener('resize', onReflow);
    return () => {
      window.removeEventListener('scroll', onReflow, true);
      window.removeEventListener('resize', onReflow);
    };
  }, [open, place]);

  // Move focus into the menu when it opens so the keyboard path continues
  // without a second tab press.
  useEffect(() => {
    if (open) enabledItems()[0]?.focus();
  }, [open]);

  const handleMenuKeyDown = (e: React.KeyboardEvent<HTMLDivElement>): void => {
    if (e.key === 'Escape') {
      e.stopPropagation();
      setOpen(false);
      return;
    }
    if (e.key !== 'ArrowDown' && e.key !== 'ArrowUp') return;
    e.preventDefault();
    const items = enabledItems();
    if (items.length === 0) return;
    const at = items.indexOf(document.activeElement as HTMLButtonElement);
    const next = e.key === 'ArrowDown' ? at + 1 : at - 1;
    items[(next + items.length) % items.length]?.focus();
  };

  return (
    <div ref={rootRef} className="relative shrink-0">
      <button
        ref={triggerRef}
        type="button"
        // The card behind this is itself clickable, so a press here must not
        // also open the task.
        onClick={(e) => {
          e.stopPropagation();
          setOpen((v) => !v);
        }}
        onPointerDown={(e) => e.stopPropagation()}
        disabled={disabled}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={open ? menuId : undefined}
        aria-label={t('flow.board.moveTask', { title: taskTitle })}
        title={t('flow.board.moveTask', { title: taskTitle })}
        className="relative flex h-6 w-6 items-center justify-center rounded-sm hover:bg-[var(--flow-hover)] disabled:opacity-40 after:absolute after:-inset-2 after:content-['']"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        <CornerUpRight size={12} aria-hidden />
      </button>

      {open &&
        createPortal(
          <div
            ref={menuRef}
            id={menuId}
            role="menu"
            aria-label={t('flow.board.moveTaskMenu')}
            onKeyDown={handleMenuKeyDown}
            className="fixed z-[60] min-w-[10rem] rounded-lg py-1 shadow-lg"
            style={{
              background: 'var(--flow-card)',
              border: '1px solid var(--flow-border)',
              top: position?.top ?? 0,
              left: position?.left ?? 0,
              // Avoid a frame at the wrong spot before the measurement lands.
              visibility: position === null ? 'hidden' : 'visible',
            }}
          >
            {targets.length === 0 ? (
              <p className="px-3 py-2 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                {t('flow.board.moveNoTargets')}
              </p>
            ) : (
              targets.map((target) => {
                const isCurrent = target.id === currentColumnId;
                return (
                  <button
                    key={target.id}
                    data-move-item
                    type="button"
                    role="menuitem"
                    aria-current={isCurrent || undefined}
                    disabled={isCurrent}
                    onClick={(e) => {
                      e.stopPropagation();
                      setOpen(false);
                      onMove(target.id);
                    }}
                    onPointerDown={(e) => e.stopPropagation()}
                    className="flex w-full min-h-[44px] items-center gap-2 px-3 text-start text-meta hover:bg-[var(--flow-hover)] disabled:opacity-50"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {isCurrent ? (
                      <Check size={12} aria-hidden />
                    ) : (
                      <span className="w-3" aria-hidden />
                    )}
                    <span className="truncate">{target.name}</span>
                  </button>
                );
              })
            )}
          </div>,
          document.body
        )}
    </div>
  );
};

export default MoveToColumnMenu;
