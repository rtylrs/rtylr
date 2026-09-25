import React from 'react';

export const FullScreenLoader: React.FC = () => {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-admin-bg">
      <div className="w-8 h-8 border-2 border-black/10 border-t-admin rounded-full animate-spin" />
    </div>
  );
};
