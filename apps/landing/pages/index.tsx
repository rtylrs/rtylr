import React from 'react';
import HelmetContainer from '../components/HelmetContainer';
import styles from './index.module.scss';

const HomePage = (): JSX.Element => {
  return (
    <>
      <HelmetContainer title="Rtylr | Run your business with AI." />
      <div className={styles.comingSoon}>
        <div className={styles.content}>
          <h1 className={styles.title}>Coming Soon</h1>
        </div>
      </div>
    </>
  );
};

export default HomePage;
