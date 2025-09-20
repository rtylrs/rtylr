import React, { useMemo } from 'react';
import Header from './Header';
import Footer from './Footer/Footer';

const MainLayout: React.FC<{
  children: React.ReactNode;
  hasPadding?: boolean;
  paddingTop?: string;
  hasHeader?: boolean;
  hasFooter?: boolean;
}> = React.memo(
  ({
    children,
    hasPadding = true,
    paddingTop = '30px',
    hasHeader = true,
    hasFooter = true,
  }) => {
    const marginTop = useMemo(() => {
      if (typeof window === 'undefined') return '0px';
      const pathname = window.location.pathname;
      if (pathname === '/') {
        return '0px';
      }
      return hasPadding ? paddingTop : '0px';
    }, [hasPadding, paddingTop]);

    const contentStyle = useMemo(
      () => ({
        marginTop,
      }),
      [marginTop]
    );

    return (
      <div>
        {hasHeader && <Header />}
        <div style={contentStyle}>{children}</div>
        {hasFooter && <Footer />}
      </div>
    );
  }
);

MainLayout.displayName = 'MainLayout';

export default MainLayout;
