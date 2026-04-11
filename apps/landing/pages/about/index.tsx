import React from 'react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import { useConfigStore } from '../../stores/configStore';
import styles from './about.module.scss';

const AboutPage = (): JSX.Element => {
  const aboutContent = useConfigStore((state) => state.aboutContent);

  return (
    <AppLayout>
      <MainLayout>
        <div className={styles['about-container']}>
          {aboutContent ? (
            <div dangerouslySetInnerHTML={{ __html: aboutContent }} />
          ) : (
            <div className={styles['about-fallback']}>
              <h1>About Rtylr</h1>
              <p>Welcome to Rtylr - transforming restaurant experiences with digital innovation.</p>
            </div>
          )}
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default AboutPage;
