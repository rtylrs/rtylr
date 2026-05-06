import React from 'react';

export const FullScreenLoader: React.FC = () => {
  return (
    <div className="fixed inset-0 flex items-center justify-center bg-pos-bg z-50">
      <div className="w-10 h-10 border-4 border-white/20 border-t-pos-text rounded-full animate-spin" />
    </div>
  );
};
