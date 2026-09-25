import React from 'react';

interface CustomSidebarLayoutProps {
  children?: React.ReactNode;
  sidebar?: React.ReactNode;
  navbar?: React.ReactNode;
  showNavbar?: boolean;
}

const CustomSidebarLayout: React.FC<CustomSidebarLayoutProps> = ({
  children,
  sidebar,
  navbar,
  showNavbar = true,
}) => {
  return (
    <div className="flex min-h-screen bg-admin-bg">
      {showNavbar && navbar && <div className="fixed top-0 left-0 right-0 z-30">{navbar}</div>}
      {sidebar && <div className="w-60 shrink-0 fixed top-0 left-0 h-screen z-20">{sidebar}</div>}
      <div className={`flex-1 ${sidebar ? 'ml-60' : ''} ${showNavbar && navbar ? 'mt-14' : ''}`}>
        {children}
      </div>
    </div>
  );
};

export default CustomSidebarLayout;
