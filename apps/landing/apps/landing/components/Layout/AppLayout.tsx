import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';

import CustomScrollbar from '../CustomScrollbar/CustomScrollbar';

// Import Zustand stores and React Query hooks
import { useUserStore } from '../../stores/userStore';
import { useUserInfo } from '../../hooks/api/useUserApi';

interface AppLayoutProps {
  children: React.ReactNode;
}

const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const router = useRouter();

  // Zustand stores
  const { isAuthenticated, role } = useUserStore();

  // React Query hooks - only on client side
  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    setIsClient(true);
  }, []);

  const { data: userData } = useUserInfo();

  // States


  return (
    <React.Fragment>
      {/* Toolbar and Cookies components temporarily commented out */}
      {/* {(role && role === 'admin') ||
      (typeof window !== 'undefined' && window.location.hostname === 'localhost') ? (
        <ToolBar />
      ) : null} */}

      {/* <Cookies /> */}

      {children}
      <CustomScrollbar />
    </React.Fragment>
  );
};

export default AppLayout;
