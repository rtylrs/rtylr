import React from 'react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import HeroSection from './components/HeroSection/HeroSection';
import PricingSection from './components/PricingSection/PricingSection';
import styles from './Home.module.scss';

const NewHomePage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer title="Rtylr | Run your business with AI." />
        <div className={styles.newHomepage}>
          <HeroSection />
          <PricingSection />
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default NewHomePage;
