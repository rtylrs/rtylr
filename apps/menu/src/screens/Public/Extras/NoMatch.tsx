import React from 'react';
import { HelmetContainer } from '@voxire/components';

const NoMatch: React.FC = () => {
  return (
    <>
      <HelmetContainer title="Page not found · rtylr" />
      <div className="min-h-screen flex flex-col items-center justify-center gap-3 bg-white">
        <span className="text-7xl font-black text-black/10">404</span>
        <p className="text-sm text-gray-500">Page not found</p>
      </div>
    </>
  );
};

export default NoMatch;
