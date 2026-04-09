import React from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './NoMatch.module.scss';

const NoMatch: React.FC = () => {
  const navigate = useNavigate();
  return (
    <div className={styles.page}>
      <h1 className={styles.code}>404</h1>
      <p className={styles.msg}>Page not found</p>
      <button className={styles.btn} onClick={() => navigate('/')}>
        Back to POS
      </button>
    </div>
  );
};

export default NoMatch;
