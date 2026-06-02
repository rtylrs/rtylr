import React from 'react';
import styles from './GridLines.module.scss';

/**
 * Full-page grid lines, Retool-style vertical and horizontal rules
 * that give the page a structured, architectural feel.
 * Placed once in MainLayout, rendered fixed behind all content.
 */
const GridLines: React.FC = (): React.JSX.Element => {
  return (
    <div className={styles.grid} aria-hidden="true">
      {/* Vertical lines */}
      <span className={styles.vLine} style={{ left: '25%' }} />
      <span className={styles.vLine} style={{ left: '50%' }} />
      <span className={styles.vLine} style={{ left: '75%' }} />
    </div>
  );
};

export default GridLines;
