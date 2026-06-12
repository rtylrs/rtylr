import React, { useEffect, useRef } from 'react';
import { useInView, useMotionValue, useSpring } from 'motion/react';

interface CountUpProps {
  value: string;
  className?: string;
}

// Animates the numeric portion of a stat ("50K+", "99.9%", "24/7") from 0
// to its final value with a spring when it scrolls into view. Values with
// no leading number render as-is.
export const CountUp: React.FC<CountUpProps> = ({ value, className }): React.JSX.Element => {
  const match = /^(\d+(?:\.\d+)?)(.*)$/.exec(value);
  const target = match ? parseFloat(match[1]) : null;
  const suffix = match ? match[2] : '';
  const decimals = match && match[1].includes('.') ? match[1].split('.')[1].length : 0;

  const ref = useRef<HTMLSpanElement>(null);
  const inView = useInView(ref, { once: true, margin: '-60px' });
  const motionValue = useMotionValue(0);
  const spring = useSpring(motionValue, { stiffness: 90, damping: 28, mass: 1 });

  useEffect(() => {
    if (inView && target !== null) motionValue.set(target);
  }, [inView, target, motionValue]);

  useEffect(() => {
    if (target === null) return undefined;
    return spring.on('change', (latest) => {
      if (ref.current) {
        ref.current.textContent = `${latest.toFixed(decimals)}${suffix}`;
      }
    });
  }, [spring, target, decimals, suffix]);

  if (target === null) {
    return <span className={className}>{value}</span>;
  }

  return (
    <span ref={ref} className={className}>
      {`0${suffix}`}
    </span>
  );
};
