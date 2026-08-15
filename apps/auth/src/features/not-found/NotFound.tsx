import React from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';

export const NotFound: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  return (
    <>
      <HelmetContainer title="404 - Page Not Found · rtylr" />
      <div className="flex flex-col items-center justify-center min-h-[60vh] text-center p-6 space-y-4">
        <div className="w-16 h-16 rounded-2xl bg-amber-500/10 text-amber-600 flex items-center justify-center font-bold text-2xl">
          404
        </div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          {t('auth.notFound.title', 'Page Not Found')}
        </h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
          {t('auth.notFound.subtitle', 'The requested page does not exist or has been moved.')}
        </p>
        <button
          type="button"
          onClick={() => navigate({ to: '/' })}
          className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-lg bg-black text-white dark:bg-white dark:text-black hover:opacity-90 transition-opacity min-h-[44px]"
        >
          {t('auth.notFound.returnToSignIn', 'Return to Sign In')}
        </button>
      </div>
    </>
  );
};

export default NotFound;
