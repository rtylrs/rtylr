import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import {
  Chart as ChartJS,
  LineElement,
  PointElement,
  LinearScale,
  Title,
  Tooltip,
  Legend,
  CategoryScale,
} from 'chart.js';

import i18next from '../../services/i18next';
import HelmetContainer from '../HelmetContainer';
import { getCookie } from '../../lib/cookies';
import CustomScrollbar from '../CustomScrollbar/CustomScrollbar';

// Import Zustand stores and React Query hooks
import { useUserStore } from '../../stores/userStore';
import { useConfigStore } from '../../stores/configStore';
import { useConfigData } from '../../hooks/api/useConfigApi';
import { useUserInfo } from '../../hooks/api/useUserApi';

ChartJS.register(LineElement, PointElement, LinearScale, Title, Tooltip, Legend, CategoryScale);

interface AppLayoutProps {
  children: React.ReactNode;
}

const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const router = useRouter();

  // Zustand stores
  const { isAuthenticated, role } = useUserStore();
  const configStore = useConfigStore();

  // React Query hooks - only on client side
  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    setIsClient(true);
  }, []);

  const { data: configData } = useConfigData();
  const { data: userData } = useUserInfo();

  // States

  useEffect(() => {
    const language = getCookie('language');
    if (language) {
      i18next.changeLanguage(language);
    }
  }, []);

  return (
    <React.Fragment>
      {configData && (
        <HelmetContainer title={'Rtylr - Contactless QR Menu'} favicon={'/favicon.ico'} />
      )}

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
