import React from 'react';
import AppLayout from '../components/Layout/AppLayout';
import MainLayout from '../layouts/Main/MainLayout';
import HelmetContainer from '../components/HelmetContainer';
// Section components live under components/home/ (not pages/home/) so
// Next.js's file-system router doesn't generate a route for every
// subfolder — that was dumping /home/HeroSection/... routes into the
// sitemap and exposing thin-content duplicates of the homepage.
import HeroSection from '../components/home/HeroSection/HeroSection';
import ProductsSection from '../components/home/ProductsSection/ProductsSection';
import ParallaxShowcase from '../components/home/ParallaxShowcase/ParallaxShowcase';
import FeaturesSection from '../components/home/FeaturesSection/FeaturesSection';
import HowItWorksSection from '../components/home/HowItWorksSection/HowItWorksSection';
import PricingSection from '../components/home/PricingSection/PricingSection';
import TestimonialsSection from '../components/home/TestimonialsSection/TestimonialsSection';
import IntegrationsSection from '../components/home/IntegrationsSection/IntegrationsSection';
import CTASection from '../components/home/CTASection/CTASection';
import styles from './index.module.scss';

const HomePage = (): JSX.Element => {
  const baseUrl = 'https://rtylr.com';
  const description =
    'POS, ERP, HR, CRM, Finance & Flow. The all-in-one business OS built for modern teams of every size. Powered by AI.';
  const title = 'rtylr · Your entire business, one platform.';
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
          keywords="POS, ERP, HR, CRM, Finance, Flow, business OS, inventory management, payroll, loyalty programs, invoicing, project management, restaurant software, retail software, AI"
        />
        <div className={styles.Homepage}>
          <HeroSection />
          <ProductsSection />
          <ParallaxShowcase />
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
};

export default HomePage;
