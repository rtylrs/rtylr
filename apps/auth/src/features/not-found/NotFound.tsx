import React from 'react';
import { HelmetContainer } from '@voxire/components';

export const NotFound: React.FC = () => {
  return (
    <>
      <HelmetContainer title="404 - Page Not Found · rtylr" />
      <div className="flex flex-col items-center justify-center min-h-[60vh] text-center p-6 space-y-4">
        <div className="w-16 h-16 rounded-2xl bg-amber-500/10 text-amber-600 flex items-center justify-center font-bold text-2xl">
          404
        </div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Page Not Found</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
          The requested page does not exist or has been moved.
        </p>
        <a
          href="/"
          className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-lg bg-black text-white dark:bg-white dark:text-black hover:opacity-90 transition-opacity"
        >
          Return to Sign In
        </a>
      </div>
    </>
  );
};

export default NotFound;
