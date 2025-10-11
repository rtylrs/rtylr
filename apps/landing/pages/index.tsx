import React from 'react';
import AppLayout from '../components/Layout/AppLayout';
import MainLayout from '../layouts/Main/MainLayout';
import HelmetContainer from '../components/HelmetContainer';
import HeroSection from './home/HeroSection/HeroSection';
import PricingSection from './home/PricingSection/PricingSection';
import styles from './index.module.scss';

const HomePage = (): JSX.Element => {
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

export default HomePage;
