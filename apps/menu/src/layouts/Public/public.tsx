import React, { useMemo } from 'react';
import Header from './PublicHeader';
import Footer from './PublicFooter/PublicFooter';

const Public: React.FC<{
  children: React.ReactNode;
  hasPadding?: boolean;
  paddingTop?: string;
  hasHeader?: boolean;
  hasFooter?: boolean;
}> = React.memo(
  ({ children, hasPadding = true, paddingTop = '30px', hasHeader = true, hasFooter = true }) => {
    const marginTop = useMemo(() => {
      const pathname = document.location.pathname;
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

Public.displayName = 'Public';

export default Public;
