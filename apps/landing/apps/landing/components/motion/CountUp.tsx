import React, { useEffect, useRef } from 'react';
import { useInView, useMotionValue, useSpring } from 'motion/react';

interface CountUpProps {
  value: string;
  className?: string;
}

// Animates the numeric portion of a stat ("50K+", "99.9%", "24/7",
// "99,9 %") from 0 to its final value with a spring when it scrolls into
// view. Values with no leading number render as-is.
//
// The decimal separator is captured, not assumed: French writes 99,9 as
// a comma, and toFixed() always emits a period regardless of locale, so
// the separator from the source string is swapped back in on every
// animation frame rather than hardcoding one.
export const CountUp: React.FC<CountUpProps> = ({ value, className }): React.JSX.Element => {
  const match = /^(\d+)(?:([.,])(\d+))?(.*)$/.exec(value);
  const decimalSeparator = match?.[2] ?? '.';
  const decimalDigits = match?.[3] ?? '';
  const target = match ? parseFloat(`${match[1]}.${decimalDigits || '0'}`) : null;
  const suffix = match ? match[4] : '';
  const decimals = decimalDigits.length;

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
        const formatted = latest.toFixed(decimals).replace('.', decimalSeparator);
        ref.current.textContent = `${formatted}${suffix}`;
      }
    });
  }, [spring, target, decimals, decimalSeparator, suffix]);

  if (target === null) {
    return <span className={className}>{value}</span>;
  }

  return (
    <span ref={ref} className={className}>
      {`0${suffix}`}
    </span>
  );
};
