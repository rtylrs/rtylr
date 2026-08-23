import React from 'react';
import AppLayout from '../components/Layout/AppLayout';
import MainLayout from '../layouts/Main/MainLayout';
import HelmetContainer from '../components/HelmetContainer';
import DirectAnswer from '../components/DirectAnswer/DirectAnswer';
import FaqSection from '../components/FaqSection/FaqSection';
import ProductTruthSection from '../components/home/ProductTruthSection/ProductTruthSection';
import { useFaq } from '../lib/faqs';
import { webSiteSchema, organizationSchema, productSchema } from '../lib/structured-data';
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
  // SoftwareApplication belongs here, on the page that describes the
  // product. It used to arrive from _document.tsx, which put it on all
  // 24 pages including the privacy policy; this page is the one place
  // sitewide where it was actually true.
  structuredData: {
    '@context': 'https://schema.org',
    '@graph': [
      webSiteSchema(),
      organizationSchema(),
      productSchema({
        name: 'rtylr',
        path: '/',
        description:
          'One business platform covering POS, ERP, CRM, HR, Finance and Flow on a single data model, so a sale updates stock and the books without an export.',
      }),
    ],
  },
};

const WebsiteContent = (): React.JSX.Element => {
  const faq = useFaq('home');

  return (
    <AppLayout>
      <MainLayout>
        <div className={styles.Homepage}>
          <HeroSection />
          <DirectAnswer label="What is rtylr?">
            rtylr is an all-in-one business operating system for small and mid-sized companies:
            point of sale, ERP, HR, CRM, finance and project management on a single database. It is
            built for restaurants, retail and services businesses running 2 to 200 staff, across one
            location or twenty. The point is not that it is seven apps. It is that a sale moves
            stock, revenue and loyalty without anyone exporting a file.
          </DirectAnswer>
          <ProductsSection />
          <FeaturesSection />
          <HowItWorksSection />
          <PricingSection />
          <TestimonialsSection />
          {/* Verifiable claims, sitting directly after the quotes. See the
            component for what was cut from it and why. */}
          <ProductTruthSection />
          <IntegrationsSection />
          <FaqSection title={faq.title} intro={faq.intro} faqs={faq.items} />
          <CTASection />
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export { WebsiteContent };

const HomePage = (): React.JSX.Element => {
  return (
    <>
      <HelmetContainer {...SEO_PROPS} />
      <WebsiteContent />
    </>
  );
};

export default HomePage;
