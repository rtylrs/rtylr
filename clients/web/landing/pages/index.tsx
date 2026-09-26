import React from 'react';
import AppLayout from '../components/Layout/AppLayout';
import MainLayout from '../layouts/Main/MainLayout';
import HelmetContainer from '../components/HelmetContainer';
import DirectAnswer from '../components/DirectAnswer/DirectAnswer';
import FaqSection from '../components/FaqSection/FaqSection';
import ProductTruthSection from '../components/home/ProductTruthSection/ProductTruthSection';
import { useTranslation } from 'react-i18next';
import { useFaq } from '../lib/faqs';
import { webSiteSchema, organizationSchema, productSchema } from '../lib/structured-data';
import HeroSection from '../components/home/HeroSection/HeroSection';
import ProductsSection from '../components/home/ProductsSection/ProductsSection';
import FeaturesSection from '../components/home/FeaturesSection/FeaturesSection';
import HowItWorksSection from '../components/home/HowItWorksSection/HowItWorksSection';
import PricingSection from '../components/home/PricingSection/PricingSection';
import IntegrationsSection from '../components/home/IntegrationsSection/IntegrationsSection';
import CTASection from '../components/home/CTASection/CTASection';
import styles from './index.module.scss';

/**
 * Built inside the component, not at module scope: t() needs the hook. The
 * non-copy half (url, ogType, ogLocale, twitterCard) stays static.
 */
function useSeoProps() {
  const { t } = useTranslation();
  return {
    title: t('home.meta.title'),
    description: t('home.meta.description'),
    url: 'https://rtylr.com/',
    ogType: 'website' as const,
    ogSiteName: t('home.meta.ogSiteName'),
    ogDescription: t('home.meta.description'),
    twitterCard: 'summary_large_image' as const,
    keywords: t('home.meta.keywords'),
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
          description: t('home.meta.schemaDescription'),
        }),
      ],
    },
  };
}

const WebsiteContent = (): React.JSX.Element => {
  const faq = useFaq('home');
  const { t } = useTranslation();

  return (
    <AppLayout>
      <MainLayout>
        <div className={styles.Homepage}>
          <HeroSection />
          <DirectAnswer label={t('home.directAnswer.label')}>
            {t('home.directAnswer.body')}
          </DirectAnswer>
          <ProductsSection />
          <FeaturesSection />
          <HowItWorksSection />
          <PricingSection />
          {/* Verifiable claims. See the component for what is in it and why. */}
          <ProductTruthSection />
          <IntegrationsSection />
          <FaqSection title={faq.title} intro={faq.intro} faqs={faq.items} />
          <CTASection />
        </div>
      </MainLayout>
    </AppLayout>
  );
};

const HomePage = (): React.JSX.Element => {
  const seo = useSeoProps();

  return (
    <>
      <HelmetContainer {...seo} />
      <WebsiteContent />
    </>
  );
};

export default HomePage;
