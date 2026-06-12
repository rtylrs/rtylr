import React, { useMemo } from 'react';
import Header from './Header';

const Public: React.FC<{
  children: React.ReactNode;
  hasPadding?: boolean;
  paddingTop?: string;
  hasHeader?: boolean;
}> = React.memo(({ children, hasPadding = true, paddingTop = '30px', hasHeader = true }) => {
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
    </div>
  );
});

export default Public;
