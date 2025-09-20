import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { motion } from 'framer-motion';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faChevronUp } from '@fortawesome/free-solid-svg-icons';
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
import Maintanance from '../../screens/Maintanance';
import HelmetContainer from '../HelmetContainer';
import { getCookie } from '../../lib/cookies';
import useColors from '../../hooks/useColors';

// Import Zustand stores and React Query hooks
import { useUserStore } from '../../stores/userStore';
import { useConfigStore } from '../../stores/configStore';
import { useConfigData } from '../../hooks/api/useConfigApi';
import { useUserInfo } from '../../hooks/api/useUserApi';

ChartJS.register(
  LineElement,
  PointElement,
  LinearScale,
  Title,
  Tooltip,
  Legend,
  CategoryScale
);

interface AppLayoutProps {
  children: React.ReactNode;
}

const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const router = useRouter();
  const { primary } = useColors();

  // Zustand stores
  const { isAuthenticated, role } = useUserStore();
  const { maintanance } = useConfigStore();

  // React Query hooks - only on client side
  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    setIsClient(true);
  }, []);

  const { data: configData } = useConfigData();
  const { data: userData } = useUserInfo();

  // States
  const [showTopBtn, setShowTopBtn] = useState(false);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handleScroll = () => {
      if (window.scrollY > 400) {
        setShowTopBtn(true);
      } else {
        setShowTopBtn(false);
      }
    };

    window.addEventListener('scroll', handleScroll);

    return () => {
      window.removeEventListener('scroll', handleScroll);
    };
  }, []);

  useEffect(() => {
    const language = getCookie('language');
    if (language) {
      i18next.changeLanguage(language);
    }
  }, []);

  return (
    <React.Fragment>
      {configData && (
        <HelmetContainer
          title={'Vbitez - Contactless QR Menu'}
          favicon={'/favicon.ico'}
        />
      )}

      {maintanance &&
        !router.pathname.includes('/manage/') &&
        (!role || role !== 'admin') && <Maintanance />}

      {/* Toolbar and Cookies components temporarily commented out */}
      {/* {(role && role === 'admin') ||
      (typeof window !== 'undefined' && window.location.hostname === 'localhost') ? (
        <ToolBar />
      ) : null} */}

      {/* <Cookies /> */}

      {showTopBtn && !router.pathname.includes('/restaurant/') && (
        <motion.div
          className="ScrollToTop"
          onClick={() => {
            setShowTopBtn(false);
            if (typeof window !== 'undefined') {
              window.scrollTo({
                top: 0,
                behavior: 'smooth',
              });
            }
          }}
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          whileHover={{ scale: 1.2 }}
          whileTap={{ scale: 0.8 }}
          style={{
            backgroundColor: primary,
          }}
        >
          <FontAwesomeIcon icon={faChevronUp} />
        </motion.div>
      )}

      {children}
    </React.Fragment>
  );
};

export default AppLayout;