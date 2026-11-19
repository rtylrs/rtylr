import React from 'react';

import { cn } from '../../lib/cn';
import type { Cover as CoverData } from '../../lib/blog/types';
import styles from './Cover.module.scss';

/**
 * A post's cover: a small product-style panel drawn in markup, the way the
 * rest of the site uses product UI as its imagery. No image file, so it
 * costs no request, mirrors in RTL and stays sharp at any size.
 *
 * Decorative. The post title carries the meaning, so the panel is hidden
 * from assistive technology.
 */
const BARS = [38, 54, 46, 70, 62, 84, 76];

export function Cover({
  cover,
  size = 'card',
}: {
  cover: CoverData;
  size?: 'card' | 'hero';
}): React.JSX.Element {
  return (
    <div className={cn(styles.frame, styles[size])} aria-hidden="true">
      <div className={styles.panel}>
        <div className={styles.bar}>
          <span className={styles.dots}>
            <i />
            <i />
            <i />
          </span>
          <span className={styles.title}>{cover.title}</span>
        </div>
        {cover.motif === 'chart' && (
          <div className={styles.chart}>
            {BARS.map((h, i) => (
              <span key={i} style={{ height: `${h}%` }} />
            ))}
          </div>
        )}
        {cover.motif === 'board' && (
          <div className={styles.board}>
            {[3, 2, 1].map((n, col) => (
              <span key={col} className={styles.column}>
                {Array.from({ length: n }, (_, i) => (
                  <i key={i} />
                ))}
              </span>
            ))}
          </div>
        )}
        {cover.motif === 'schedule' && (
          <div className={styles.schedule}>
            {[
              [0, 55],
              [15, 60],
              [30, 50],
            ].map(([start, width], i) => (
              <span key={i} style={{ insetInlineStart: `${start}%`, width: `${width}%` }} />
            ))}
          </div>
        )}
        {cover.motif === 'caret' && <div className={styles.prompt}>{cover.rows[0]?.label}</div>}
        <ul className={styles.rows}>
          {(cover.motif === 'caret' ? cover.rows.slice(1) : cover.rows).map((row, i) => (
            <li key={i}>
              <span className={styles.label}>{row.label}</span>
              {row.value && (
                <span className={cn(styles.value, row.tone && styles[row.tone])}>{row.value}</span>
              )}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

export default Cover;
