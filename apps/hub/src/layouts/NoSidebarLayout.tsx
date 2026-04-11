import React from 'react';

interface NoSidebarLayoutProps {
  children?: React.ReactNode;
}

const NoSidebarLayout: React.FC<NoSidebarLayoutProps> = ({ children }) => {
  return (
    <div className="flex flex-col min-h-screen" style={{ background: '#0a0a0a' }}>
      <div className="flex-1 p-6">{children}</div>
    </div>
  );
};

export default NoSidebarLayout;
