import React from 'react';
import { useTranslation } from 'react-i18next';
import { useRouter } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { SearchX } from 'lucide-react';

const NoMatch = (): React.JSX.Element => {
  const { t } = useTranslation('notfound');
  const router = useRouter();

  return (
    <>
      <HelmetContainer title="Page not found · rtylr" />
      <div className="min-h-screen flex items-center justify-center bg-white">
        <EmptyState
          icon={SearchX}
          title={t('notfound.title')}
          description={t('notfound.description')}
          action={{ label: t('notfound.goBack'), onClick: () => router.history.back() }}
        />
      </div>
    </>
  );
};

export default NoMatch;
