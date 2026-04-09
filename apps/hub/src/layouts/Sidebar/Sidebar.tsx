import React from 'react';
import { SidebarHeader, SidebarUser, SidebarMenu } from './components';

const Sidebar = () => (
  <div
    className="flex flex-col h-full bg-admin overflow-hidden"
    style={{ gridArea: 'sidebar' }}
  >
    <SidebarHeader />
    <SidebarMenu />
    <SidebarUser />
  </div>
);

export default Sidebar;
