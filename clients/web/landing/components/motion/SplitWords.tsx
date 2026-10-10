import React from 'react';
import { motion } from 'motion/react';
import { EASE_OUT_EXPO } from '../../lib/animations';

interface SplitWordsProps {
  text: string;
  as?: 'h1' | 'h2' | 'h3' | 'p' | 'span';
  className?: string;
  delay?: number;
  stagger?: number;
  duration?: number;
  // 'mount' animates immediately; 'view' waits for the element to scroll in
  trigger?: 'mount' | 'view';
  // Applied to every line after the first, e.g. to mute a follow-up line
  // for the two-tone headline style.
  restLinesClassName?: string;
}

// Per-word masked rise. Each word sits inside an overflow-hidden clip and
// climbs into place, so headlines assemble line-by-line instead of fading
// in as one block.
export const SplitWords: React.FC<SplitWordsProps> = ({
  text,
  as = 'span',
  className,
  delay = 0,
  stagger = 0.055,
  duration = 0.9,
  trigger = 'mount',
  restLinesClassName,
}): React.JSX.Element => {
  const Tag = motion[as];
  const lines = text.split('\n');

  const viewProps =
    trigger === 'view'
      ? { whileInView: 'visible' as const, viewport: { once: true, margin: '-80px' } }
      : { animate: 'visible' as const };

  let wordIndex = 0;

  return (
    <Tag
      className={className}
      initial="hidden"
      {...viewProps}
      variants={{
        hidden: {},
        visible: { transition: { staggerChildren: stagger, delayChildren: delay } },
      }}
      aria-label={text.replace(/\n/g, ' ')}
    >
      {lines.map((line, li) => (
        <React.Fragment key={li}>
          {line.split(' ').map((word, wi) => {
            const key = `${word}-${wordIndex++}`;
            return (
              <React.Fragment key={key}>
                {wi > 0 && ' '}
                <span
                  className={li > 0 ? restLinesClassName : undefined}
                  style={{
                    display: 'inline-block',
                    overflow: 'hidden',
                    verticalAlign: 'bottom',
                    // Own line-height, not the (often tighter) parent's - so
                    // this clip mask has room for descenders (g, y, p, q, j)
                    // regardless of what font/line-height the heading uses.
                    lineHeight: 'normal',
                  }}
                  aria-hidden="true"
                >
                  <motion.span
                    style={{ display: 'inline-block', willChange: 'transform' }}
                    variants={{
                      hidden: { y: '115%' },
                      visible: {
                        y: '0%',
                        transition: { duration, ease: EASE_OUT_EXPO },
                      },
                    }}
                  >
                    {word}
                  </motion.span>
                </span>
              </React.Fragment>
            );
          })}
          {li < lines.length - 1 && <br />}
        </React.Fragment>
      ))}
    </Tag>
  );
};
