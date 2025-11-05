import React from 'react';
import { useWindowDimensions } from '@voxire/shared/utils';
import HeaderDesktop from './DesktopHeader/HeaderDesktop';
import HeaderMobile from './MobileHeader/HeaderMobile';

const Header: React.FC = () => {
  const { isMobile } = useWindowDimensions();

  return isMobile ? <HeaderMobile /> : <HeaderDesktop />;
};

export default Header;
