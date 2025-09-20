import React from 'react';
import AppLayout from '../../src/components/Layout/AppLayout';
import MainLayout from '../../src/layouts/Main/MainLayout';
import News from '../../src/screens/News';

const NewsPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <News />
      </MainLayout>
    </AppLayout>
  );
};

export default NewsPage;