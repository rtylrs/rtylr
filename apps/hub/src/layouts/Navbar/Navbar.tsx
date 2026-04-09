import React from 'react';
import { SearchBar } from './components';

const Navbar = () => (
  <div
    className="flex items-center justify-between px-4 bg-white border-b border-black/6 z-10"
    style={{ gridArea: 'nav' }}
  >
    <div />
    <SearchBar />
    <div />
  </div>
);

export default Navbar;
