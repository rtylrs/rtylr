import React, { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';

/** Wait before the first name shows, so sweeping the cursor across a page stays quiet. */
const OPEN_DELAY_MS = 150;
/** Once one name has shown, moving to the next avatar swaps it straight away. */
const WARM_MS = 400;
const GAP = 6;
const EDGE = 8;

const SELECTOR = '[data-avatar-name]';

interface Shown {
  name: string;
  rect: DOMRect;
}

function avatarFrom(target: EventTarget | null): HTMLElement | null {
  if (!(target instanceof Element)) return null;
  const el = target.closest<HTMLElement>(SELECTOR);
  return el?.dataset.avatarName ? el : null;
}

/**
 * One name tooltip for every avatar in the portal.
 *
 * Anything carrying `data-avatar-name` gets it: `@voxire/ui`'s Avatar sets the
 * attribute from its `name` prop, Flow's AssigneeAvatar from its own, and a
 * button wrapping an avatar can set it directly. Mounted once at the app root
 * and driven by delegated events, so the shared package needs no tooltip of
 * its own and no avatar needs a wrapper element.
 *
 * Mouse and pen hover, and keyboard focus, show it. Touch does not, because a
 * tap would leave it stuck open over whatever the tap was for.
 */
export const AvatarNameTooltip: React.FC = () => {
  const [shown, setShown] = useState<Shown | null>(null);
  const [pos, setPos] = useState<{ left: number; top: number } | null>(null);
  const tipRef = useRef<HTMLDivElement>(null);
  const current = useRef<HTMLElement | null>(null);
  const timer = useRef<number | undefined>(undefined);
  const lastHidden = useRef(0);
  const isOpen = useRef(false);

  useEffect(() => {
    const clear = (): void => {
      window.clearTimeout(timer.current);
      timer.current = undefined;
    };
    const hide = (): void => {
      clear();
      if (isOpen.current) lastHidden.current = Date.now();
      isOpen.current = false;
      current.current = null;
      setShown(null);
    };
    const show = (el: HTMLElement, immediate: boolean): void => {
      if (current.current === el) return;
      const warm = isOpen.current || Date.now() - lastHidden.current < WARM_MS;
      clear();
      current.current = el;
      const open = (): void => {
        if (current.current !== el || !el.isConnected) return;
        isOpen.current = true;
        setShown({ name: el.dataset.avatarName ?? '', rect: el.getBoundingClientRect() });
      };
      if (immediate || warm) open();
      else timer.current = window.setTimeout(open, OPEN_DELAY_MS);
    };

    const onOver = (e: PointerEvent): void => {
      if (e.pointerType === 'touch') return;
      const el = avatarFrom(e.target);
      if (el) show(el, false);
    };
    const onOut = (e: PointerEvent): void => {
      const el = current.current;
      if (!el) return;
      const next = e.relatedTarget;
      if (next instanceof Node && el.contains(next)) return;
      hide();
    };
    const onFocus = (e: FocusEvent): void => {
      const target = e.target;
      if (!(target instanceof Element)) return;
      let visible = false;
      try {
        visible = target.matches(':focus-visible');
      } catch {
        visible = false;
      }
      if (!visible) return;
      const el = avatarFrom(target) ?? target.querySelector<HTMLElement>(SELECTOR);
      if (el?.dataset.avatarName) show(el, true);
    };
    const onKey = (e: KeyboardEvent): void => {
      if (e.key === 'Escape') hide();
    };

    document.addEventListener('pointerover', onOver);
    document.addEventListener('pointerout', onOut);
    document.addEventListener('pointerdown', hide, true);
    document.addEventListener('focusin', onFocus);
    document.addEventListener('focusout', hide);
    document.addEventListener('keydown', onKey);
    window.addEventListener('scroll', hide, true);
    window.addEventListener('resize', hide);
    return () => {
      clear();
      document.removeEventListener('pointerover', onOver);
      document.removeEventListener('pointerout', onOut);
      document.removeEventListener('pointerdown', hide, true);
      document.removeEventListener('focusin', onFocus);
      document.removeEventListener('focusout', hide);
      document.removeEventListener('keydown', onKey);
      window.removeEventListener('scroll', hide, true);
      window.removeEventListener('resize', hide);
    };
  }, []);

  // Measure once rendered, then sit above the avatar, or below it when the
  // top of the screen is too close, kept inside the viewport either way.
  useLayoutEffect(() => {
    if (!shown || !tipRef.current) {
      setPos(null);
      return;
    }
    const tip = tipRef.current.getBoundingClientRect();
    const { rect } = shown;
    let top = rect.top - tip.height - GAP;
    if (top < EDGE) top = rect.bottom + GAP;
    const centred = rect.left + rect.width / 2 - tip.width / 2;
    const left = Math.max(EDGE, Math.min(centred, window.innerWidth - tip.width - EDGE));
    setPos({ left, top });
  }, [shown]);

  if (!shown || typeof document === 'undefined') return null;
  return createPortal(
    <div
      ref={tipRef}
      role="tooltip"
      data-avatar-name-tooltip=""
      className="pointer-events-none fixed max-w-64 truncate rounded-md px-2 py-1 text-xs font-medium shadow-lg"
      style={{
        left: pos?.left ?? -9999,
        top: pos?.top ?? -9999,
        opacity: pos ? 1 : 0,
        zIndex: 2147483647,
        background: 'var(--dash-text, rgb(9, 9, 11))',
        color: 'var(--dash-bg, #fff)',
      }}
    >
      {shown.name}
    </div>,
    document.body
  );
};

export default AvatarNameTooltip;
