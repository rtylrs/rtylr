import React from 'react';
import styles from './NoMatch.module.scss';

const NoMatch: React.FC = () => {
  return (
    <div className={styles.page}>
      <h1 className={styles.code}>404</h1>
      <p className={styles.msg}>Page not found</p>
    </div>
  );
};

export default NoMatch;
