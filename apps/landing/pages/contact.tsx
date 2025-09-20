import React from 'react';
import AppLayout from '../src/components/Layout/AppLayout';
import MainLayout from '../src/layouts/Main/MainLayout';
import Contact from '../src/screens/Contact';

const ContactPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <Contact />
      </MainLayout>
    </AppLayout>
  );
};

export default ContactPage;