import { useEffect, useRef, useState } from 'react';
import type { CSSProperties, PointerEvent as ReactPointerEvent } from 'react';
import './buddy-sprite.css';

export type BuddyMood =
  | 'idle'
  | 'listening'
  | 'thinking'
  | 'working'
  | 'talking'
  | 'waiting'
  | 'happy'
  | 'sleepy'
  | 'alert';

type Look = { x: number; y: number };

// Where the eyes look in moods that have somewhere to look.
const MOOD_LOOK: Partial<Record<BuddyMood, Look>> = {
  listening: { x: -0.25, y: 0.9 },
  thinking: { x: 0.65, y: -0.85 },
  working: { x: 0, y: 0.55 },
  waiting: { x: 0, y: 0.75 },
};

const prefersReducedMotion = () =>
  typeof window !== 'undefined' &&
  typeof window.matchMedia === 'function' &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches;

/**
 * Caret Buddy's character: an ink blob with a caret tuft. It breathes and
 * blinks on its own, follows the pointer with its eyes, and acts out what
 * Buddy is doing through `mood`.
 */
export default function BuddySprite({
  mood = 'idle',
  size = 28,
  track = size >= 28,
  pokeable = false,
  className,
  title,
  color,
}: {
  mood?: BuddyMood;
  size?: number;
  /** Eyes follow the pointer. Off for tiny sprites, where it reads as jitter. */
  track?: boolean;
  pokeable?: boolean;
  className?: string;
  title?: string;
  /** A body colour; empty keeps the ink look that follows the theme. */
  color?: string;
}) {
  const ref = useRef<SVGSVGElement>(null);
  const [blink, setBlink] = useState(false);
  const [poked, setPoked] = useState(0);
  const [look, setLook] = useState<Look>({ x: 0, y: 0 });
  const lastPointer = useRef(0);

  useEffect(() => {
    if (prefersReducedMotion()) return;
    let timer: ReturnType<typeof setTimeout>;
    const schedule = () => {
      timer = setTimeout(
        () => {
          setBlink(true);
          timer = setTimeout(() => {
            setBlink(false);
            schedule();
          }, 140);
        },
        2200 + Math.random() * 4200
      );
    };
    schedule();
    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    if (!track || prefersReducedMotion()) return;
    let frame = 0;
    const onMove = (event: PointerEvent) => {
      lastPointer.current = Date.now();
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(() => {
        const box = ref.current?.getBoundingClientRect();
        if (!box) return;
        const dx = event.clientX - (box.left + box.width / 2);
        const dy = event.clientY - (box.top + box.height / 2);
        const distance = Math.hypot(dx, dy) || 1;
        const reach = Math.min(1, distance / 260);
        setLook({ x: (dx / distance) * reach, y: (dy / distance) * reach });
      });
    };
    // With no pointer around, glance about now and then.
    const wander = setInterval(() => {
      if (Date.now() - lastPointer.current < 4000) return;
      setLook(
        Math.random() < 0.35
          ? { x: 0, y: 0 }
          : { x: Math.random() * 1.6 - 0.8, y: Math.random() * 1.2 - 0.6 }
      );
    }, 2600);
    window.addEventListener('pointermove', onMove);
    return () => {
      window.removeEventListener('pointermove', onMove);
      clearInterval(wander);
      cancelAnimationFrame(frame);
    };
  }, [track]);

  useEffect(() => {
    if (!poked) return;
    const timer = setTimeout(() => setPoked(0), 700);
    return () => clearTimeout(timer);
  }, [poked]);

  const eyes = MOOD_LOOK[mood] ?? look;
  const style = {
    width: size,
    height: size,
    '--bs-px': `${eyes.x * 3.2}px`,
    '--bs-py': `${eyes.y * 3.2}px`,
    '--bs-fx': `${eyes.x * 2.2}px`,
    '--bs-fy': `${eyes.y * 1.6}px`,
    ...(color ? { '--bs-tint': color } : {}),
  } as CSSProperties;
  const shownMood = poked ? 'poked' : mood;

  return (
    <svg
      ref={ref}
      className={`buddy-sprite${className ? ` ${className}` : ''}`}
      data-mood={shownMood}
      data-blink={blink && shownMood !== 'happy' && shownMood !== 'poked' ? '' : undefined}
      data-pokeable={pokeable ? '' : undefined}
      data-tinted={color ? '' : undefined}
      viewBox="0 0 100 100"
      style={style}
      role={title ? 'img' : undefined}
      aria-label={title}
      aria-hidden={title ? undefined : true}
      onPointerDown={
        pokeable
          ? (event: ReactPointerEvent<SVGSVGElement>) => {
              event.preventDefault();
              setPoked(Date.now());
            }
          : undefined
      }
    >
      <ellipse className="bs-shadow" cx="50" cy="94" rx="25" ry="3.6" />
      <g className="bs-hop">
        <g className="bs-body">
          <path className="bs-caret" d="M41 14 L50 5 L59 14" />
          <path
            className="bs-blob"
            d="M50 18 C73 18 86 34 86 57 C86 79 71 91 50 91 C29 91 14 79 14 57 C14 34 27 18 50 18 Z"
          />
          <ellipse
            className="bs-shine"
            cx="34"
            cy="34"
            rx="8"
            ry="4.6"
            transform="rotate(-24 34 34)"
          />
          <g className="bs-face">
            <g className="bs-eyes">
              <ellipse className="bs-sclera" cx="38" cy="54" rx="8" ry="10" />
              <ellipse className="bs-sclera" cx="62" cy="54" rx="8" ry="10" />
              <g className="bs-pupils">
                <circle className="bs-pupil" cx="38" cy="55" r="4.4" />
                <circle className="bs-pupil" cx="62" cy="55" r="4.4" />
                <circle className="bs-glint" cx="39.6" cy="53" r="1.3" />
                <circle className="bs-glint" cx="63.6" cy="53" r="1.3" />
              </g>
            </g>
            <path className="bs-arcs" d="M30.5 56 Q38 46 45.5 56 M54.5 56 Q62 46 69.5 56" />
            <path className="bs-squint" d="M31 50 L44 55 L31 60 M69 50 L56 55 L69 60" />
            <ellipse className="bs-cheek" cx="27.5" cy="68" rx="5.2" ry="3" />
            <ellipse className="bs-cheek" cx="72.5" cy="68" rx="5.2" ry="3" />
            <path className="bs-smile" d="M44.5 71 Q50 75.5 55.5 71" />
            <ellipse className="bs-mouth" cx="50" cy="72.5" rx="5.6" ry="4.6" />
          </g>
        </g>
      </g>
      <g className="bs-dots">
        <circle cx="74" cy="12" r="3.2" />
        <circle cx="84" cy="7" r="3.2" />
        <circle cx="94" cy="2" r="3.2" />
      </g>
      <g className="bs-zzz">
        <path d="M74 14 h7 l-7 8 h7" />
        <path d="M86 4 h5 l-5 6 h5" />
      </g>
      <text className="bs-ask" x="82" y="20">
        ?
      </text>
    </svg>
  );
}
