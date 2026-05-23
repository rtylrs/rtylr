import React from 'react';
import HeroSection from '../home/HeroSection/HeroSection';
import ProductsSection from '../home/ProductsSection/ProductsSection';
import FeaturesSection from '../home/FeaturesSection/FeaturesSection';
import PricingSection from '../home/PricingSection/PricingSection';
import CTASection from '../home/CTASection/CTASection';

export default function LandingWindow(): React.JSX.Element {
  return (
    <div
      style={{
        height: '100%',
        overflowY: 'auto',
        overflowX: 'hidden',
        background: 'var(--app-background)',
      }}
    >
      <HeroSection />
      <ProductsSection />
      <FeaturesSection />
      <PricingSection />
      <CTASection />
    </div>
  );
}
