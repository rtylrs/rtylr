import React from 'react';
import { useLayout } from '../shared/hooks/useLayout';
import DefaultLayout from './DefaultLayout';
import NoSidebarLayout from './NoSidebarLayout';
import CustomSidebarLayout from './CustomSidebarLayout';

const AppLayout: React.FC<{
  children?: React.ReactNode;
  customSidebar?: React.ReactNode;
  customNavbar?: React.ReactNode;
}> = ({ children, customSidebar, customNavbar }) => {
  const layoutType = useLayout();

  switch (layoutType) {
    case 'no-sidebar':
      return <NoSidebarLayout>{children}</NoSidebarLayout>;
    case 'custom-sidebar':
      return (
        <CustomSidebarLayout sidebar={customSidebar} navbar={customNavbar}>
          {children}
        </CustomSidebarLayout>
      );
    case 'default':
    default:
      return <DefaultLayout>{children}</DefaultLayout>;
  }
};

export default AppLayout;
