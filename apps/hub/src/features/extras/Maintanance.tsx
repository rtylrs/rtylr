import React from 'react';

const Maintanance: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-admin-bg gap-3">
      <p className="text-5xl">🔧</p>
      <h1 className="text-xl font-semibold text-primary">Under Maintenance</h1>
      <p className="text-sm text-gray-500">We'll be back shortly.</p>
    </div>
  );
};

export default Maintanance;
