import React from 'react';
import AppLayout from '../src/components/Layout/AppLayout';
import MainLayout from '../src/layouts/Main/MainLayout';
import Home from '../src/screens/Home/Home';

const HomePage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <Home />
      </MainLayout>
    </AppLayout>
  );
};

export default HomePage;