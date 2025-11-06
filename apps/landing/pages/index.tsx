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
  const baseUrl = 'https://rtylr.com';
  const description =
    'Your single source of truth for in-store, online, and back-office operations. POS · Inventory · eCommerce · Analytics works offline, scales globally.';
  const title = 'Rtylr | Run your business with AI.';
  const ogImage = `${baseUrl}/images/home/hero-banner-top-bar-bg.webp`;

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={title}
          description={description}
          image={ogImage}
          url={baseUrl}
          ogType="website"
          ogLocale="en_US"
          ogSiteName="Rtylr"
          ogDescription={description}
          twitterCard="summary_large_image"
          keywords="POS, inventory management, eCommerce, analytics, business operations, retail software"
        />
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
