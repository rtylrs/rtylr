import React from 'react';
import AppLayout from '../src/components/Layout/AppLayout';
import MainLayout from '../src/layouts/Main/MainLayout';
import FAQ from '../src/screens/FAQ';

const FaqsPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <FAQ />
      </MainLayout>
    </AppLayout>
  );
};

export default FaqsPage;