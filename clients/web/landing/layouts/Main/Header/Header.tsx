import React, { useEffect, useState } from 'react';
import { useWindowDimensions } from '@voxire/utils';
import HeaderDesktop from './DesktopHeader/HeaderDesktop';
import HeaderMobile from './MobileHeader/HeaderMobile';

const Header: React.FC = () => {
  const { isMobile } = useWindowDimensions();
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    setHydrated(true);
  }, []);

  // Keep the server and first client render identical; switch to mobile after hydration.
  return hydrated && isMobile ? <HeaderMobile /> : <HeaderDesktop />;
};

export default Header;
