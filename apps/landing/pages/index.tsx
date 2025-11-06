import React from 'react';
import AppLayout from '../components/Layout/AppLayout';
import MainLayout from '../layouts/Main/MainLayout';
import HelmetContainer from '../components/HelmetContainer';
import HeroSection from './home/HeroSection/HeroSection';
import FeaturesSection from './home/FeaturesSection/FeaturesSection';
import HowItWorksSection from './home/HowItWorksSection/HowItWorksSection';
import UseCasesSection from './home/UseCasesSection/UseCasesSection';
import PricingSection from './home/PricingSection/PricingSection';
import TestimonialsSection from './home/TestimonialsSection/TestimonialsSection';
import IntegrationsSection from './home/IntegrationsSection/IntegrationsSection';
import CTASection from './home/CTASection/CTASection';
import styles from './index.module.scss';

const HomePage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer title="Rtylr | Run your business with AI." />
        <div className={styles.Homepage}>
          <HeroSection />
          <FeaturesSection />
          <HowItWorksSection />
          <UseCasesSection />
          <PricingSection />
          <TestimonialsSection />
          <IntegrationsSection />
          <CTASection />
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default HomePage;
