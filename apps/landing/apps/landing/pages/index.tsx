import React from 'react';
import AppLayout from '../components/Layout/AppLayout';
import MainLayout from '../layouts/Main/MainLayout';
import HelmetContainer from '../components/HelmetContainer';
import { webSiteSchema, organizationSchema } from '../lib/structured-data';
import HeroSection from '../components/home/HeroSection/HeroSection';
import ProductsSection from '../components/home/ProductsSection/ProductsSection';
import FeaturesSection from '../components/home/FeaturesSection/FeaturesSection';
import HowItWorksSection from '../components/home/HowItWorksSection/HowItWorksSection';
import PricingSection from '../components/home/PricingSection/PricingSection';
import TestimonialsSection from '../components/home/TestimonialsSection/TestimonialsSection';
import IntegrationsSection from '../components/home/IntegrationsSection/IntegrationsSection';
import CTASection from '../components/home/CTASection/CTASection';
import styles from './index.module.scss';

const SEO_PROPS = {
  title: 'rtylr · Your entire business, one platform.',
  description:
    'POS, ERP, HR, CRM, Finance & Flow. The all-in-one business OS built for modern teams of every size. Powered by AI.',
  url: 'https://rtylr.com/',
  ogType: 'website' as const,
  ogLocale: 'en_US',
  ogSiteName: 'Rtylr',
  ogDescription:
    'POS, ERP, HR, CRM, Finance & Flow. The all-in-one business OS built for modern teams of every size. Powered by AI.',
  twitterCard: 'summary_large_image' as const,
  keywords:
    'POS, ERP, HR, CRM, Finance, Flow, business OS, inventory management, payroll, loyalty programs, invoicing, project management, restaurant software, retail software, AI',
  structuredData: {
    '@context': 'https://schema.org',
    '@graph': [webSiteSchema(), organizationSchema()],
  },
};

const WebsiteContent = (): JSX.Element => (
  <AppLayout>
    <MainLayout>
      <div className={styles.Homepage}>
        <HeroSection />
        <ProductsSection />
        <FeaturesSection />
        <HowItWorksSection />
        <PricingSection />
        <TestimonialsSection />
        <IntegrationsSection />
        <CTASection />
      </div>
    </MainLayout>
  </AppLayout>
);

export { WebsiteContent };

const HomePage = (): JSX.Element => {
  return (
    <>
      <HelmetContainer {...SEO_PROPS} />
      <WebsiteContent />
    </>
  );
};

export default HomePage;
