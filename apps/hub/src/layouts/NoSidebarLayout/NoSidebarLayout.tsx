import React from "react";
import Navbar from "../Navbar/Navbar";

interface NoSidebarLayoutProps {
  children?: React.ReactNode;
}

const NoSidebarLayout: React.FC<NoSidebarLayoutProps> = ({ children }) => {
  return (
    <div className="flex flex-col min-h-screen bg-admin-bg">
      <Navbar />
      <div className="flex-1 p-6">{children}</div>
    </div>
  );
};

export default NoSidebarLayout;
