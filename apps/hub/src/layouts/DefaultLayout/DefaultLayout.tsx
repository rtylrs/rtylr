import React from 'react';
import Navbar from '../Navbar/Navbar';
import Sidebar from '../Sidebar/Sidebar';

interface DefaultLayoutProps { children?: React.ReactNode; }

const DefaultLayout: React.FC<DefaultLayoutProps> = ({ children }) => (
  <div className="grid h-screen overflow-hidden" style={{ gridTemplateColumns: '220px 1fr', gridTemplateRows: '50px 1fr', gridTemplateAreas: '"sidebar nav" "sidebar main"' }}>
    <Sidebar />
    <Navbar />
    <main className="overflow-y-auto p-6 bg-admin-bg" style={{ gridArea: 'main' }}>
      {children}
    </main>
  </div>
);

export default DefaultLayout;
