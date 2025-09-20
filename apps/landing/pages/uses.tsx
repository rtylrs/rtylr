import React from 'react';
import AppLayout from '../src/components/Layout/AppLayout';
import MainLayout from '../src/layouts/Main/MainLayout';

const UsesPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <div style={{ padding: '2rem', textAlign: 'center' }}>
          <h1>Use Cases</h1>
          <p>Discover how Vbitez can transform your restaurant operations.</p>
          <div style={{ marginTop: '2rem' }}>
            <h2>Key Use Cases</h2>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '2rem', marginTop: '2rem' }}>
              <div style={{ border: '1px solid #ddd', padding: '1.5rem', borderRadius: '8px' }}>
                <h3>Digital Menu</h3>
                <p>Replace traditional paper menus with interactive digital menus that customers can access via QR codes.</p>
              </div>
              <div style={{ border: '1px solid #ddd', padding: '1.5rem', borderRadius: '8px' }}>
                <h3>Order Management</h3>
                <p>Streamline order processing and reduce wait times with our efficient order management system.</p>
              </div>
              <div style={{ border: '1px solid #ddd', padding: '1.5rem', borderRadius: '8px' }}>
                <h3>Analytics</h3>
                <p>Gain insights into customer preferences and optimize your menu offerings with detailed analytics.</p>
              </div>
            </div>
          </div>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default UsesPage;