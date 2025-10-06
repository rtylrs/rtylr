import React from 'react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import styles from './home.module.scss';

const NewHomePage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer title="Rtylr | Run your business with AI." />
        <div className={styles.newHomepage}>
          <div className={styles.content}></div>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default NewHomePage;
