import React from 'react';
import Image from 'next/image';
import HelmetContainer from '../components/HelmetContainer';
import styles from './index.module.scss';

const HomePage = (): JSX.Element => {
  return (
    <>
      <HelmetContainer title="Rtylr | Run your business with AI." />
      <div className={styles.comingSoon}>
        <div className={styles.content}>
          <div className={styles.logoContainer}>
            <Image
              src="/logo.png"
              alt="Rtylr Logo"
              width={200}
              height={200}
              priority
              className={styles.logo}
            />
          </div>
          <h1 className={styles.title}>Coming Soon</h1>
        </div>
      </div>
    </>
  );
};

export default HomePage;
