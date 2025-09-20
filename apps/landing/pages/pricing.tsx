import React from 'react';
import AppLayout from '../src/components/Layout/AppLayout';
import MainLayout from '../src/layouts/Main/MainLayout';
import Pricing from '../src/screens/Pricing';

const PricingPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <Pricing />
      </MainLayout>
    </AppLayout>
  );
};

export default PricingPage;