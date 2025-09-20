import React from 'react';
import AppLayout from '../src/components/Layout/AppLayout';
import MainLayout from '../src/layouts/Main/MainLayout';

const RestaurantsPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <div style={{ padding: '2rem', textAlign: 'center' }}>
          <h1>Restaurants</h1>
          <p>Discover restaurants using Vbitez digital menu system.</p>
          <div style={{ marginTop: '2rem' }}>
            <h2>Featured Restaurants</h2>
            <p>Coming soon...</p>
          </div>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default RestaurantsPage;