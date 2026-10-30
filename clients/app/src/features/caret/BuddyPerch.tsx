import { useEffect, useRef, useState } from 'react';
import BuddySprite from './BuddySprite';
import type { BuddyMood } from './BuddySprite';

const STEP_PX_PER_S = 70;

/**
 * Buddy sitting on the composer's top edge. When idle it strolls to a new
 * spot now and then; otherwise it stays put and acts out its mood.
 */
export default function BuddyPerch({
  mood,
  label,
  color,
}: {
  mood: BuddyMood;
  label: string;
  color?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const [x, setX] = useState(0.86);
  const position = useRef(0.86);
  const [walk, setWalk] = useState<{ ms: number; left: boolean } | null>(null);
  const reduced =
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  useEffect(() => {
    if (reduced || mood !== 'idle') {
      setWalk(null);
      return;
    }
    let timer: ReturnType<typeof setTimeout>;
    const stroll = () => {
      timer = setTimeout(
        () => {
          const width = ref.current?.parentElement?.clientWidth ?? 600;
          const from = position.current;
          const to = 0.08 + Math.random() * 0.84;
          const ms = Math.max(600, (Math.abs(to - from) * width * 1000) / STEP_PX_PER_S);
          position.current = to;
          setWalk({ ms, left: to < from });
          setX(to);
          timer = setTimeout(() => {
            setWalk(null);
            stroll();
          }, ms);
        },
        6000 + Math.random() * 9000
      );
    };
    stroll();
    return () => clearTimeout(timer);
  }, [mood, reduced]);

  return (
    <div
      ref={ref}
      className="buddy-perch"
      data-walking={walk ? '' : undefined}
      data-facing={walk?.left ? 'left' : 'right'}
      style={{ left: `${x * 100}%`, transitionDuration: walk ? `${walk.ms}ms` : '0ms' }}
    >
      <BuddySprite mood={walk ? 'idle' : mood} size={38} pokeable title={label} color={color} />
    </div>
  );
}
