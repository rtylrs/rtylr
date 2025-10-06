import React from 'react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import Hero from './home/Components/Hero';
import Features from './home/Components/Features';
import Price from './home/Components/Price';
import Demo from './home/Components/Demo';
import HelmetContainer from '../../components/HelmetContainer';
import Ticker from '../../components/Ticker/Ticker';

const LegacyHomePage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer title="Rtylr Legacy Homepage" />
        <Hero />
        <Ticker />
        <Features />
        <Demo />
        <Price />
      </MainLayout>
    </AppLayout>
  );
};

export default LegacyHomePage;