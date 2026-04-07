import React from 'react';
import Navbar from '../Navbar/Navbar';
import Sidebar from '../Sidebar/Sidebar';

const AppLayout: React.FC<{
  children?: React.ReactNode;
}> = (props) => {
  return (
    <div className="app-layout">
      <Sidebar />
      <Navbar />
      <div className="app-layout__content">
        {props.children}
      </div>
    </div>
  );
};

export default AppLayout;