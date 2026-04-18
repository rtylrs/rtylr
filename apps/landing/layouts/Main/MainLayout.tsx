import React, { useMemo, useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Header from './Header/Header';
import Footer from './Footer/Footer';
import GridLines from '../../components/GridLines/GridLines';

const MainLayout: React.FC<{
  children: React.ReactNode;
  hasPadding?: boolean;
  paddingTop?: string;
  hasHeader?: boolean;
  hasFooter?: boolean;
}> = React.memo(
  ({ children, hasPadding = true, paddingTop = '30px', hasHeader = true, hasFooter = true }) => {
    const router = useRouter();
    const [isHomePage, setIsHomePage] = useState(true);

    useEffect(() => {
      setIsHomePage(router.pathname === '/');
    }, [router.pathname]);

    const marginTop = useMemo(() => {
      if (isHomePage) {
        return '0px';
      }
      return hasPadding ? paddingTop : '0px';
    }, [isHomePage, hasPadding, paddingTop]);

    const contentStyle = useMemo(
      () => ({
        marginTop,
      }),
      [marginTop]
    );

    return (
      <div style={{ position: 'relative' }}>
        <GridLines />
        {hasHeader && <Header />}
        <div style={contentStyle}>{children}</div>
        {hasFooter && <Footer />}
      </div>
    );
  }
);

MainLayout.displayName = 'MainLayout';

export default MainLayout;
