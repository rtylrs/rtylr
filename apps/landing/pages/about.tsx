import React from 'react';
import AppLayout from '../src/components/Layout/AppLayout';
import MainLayout from '../src/layouts/Main/MainLayout';
import About from '../src/screens/About';

const AboutPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <About />
      </MainLayout>
    </AppLayout>
  );
};

export default AboutPage;