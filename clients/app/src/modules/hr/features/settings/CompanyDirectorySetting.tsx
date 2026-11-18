import React from 'react';
import { useTranslation } from 'react-i18next';
import { Switch } from '@voxire/ui';
import { useGetHRConfig, useUpsertHRConfig } from '../../shared/hooks/queries/useSetup';
import { Card, QueryError, Skeleton } from '../../shared/components/ui';

/**
 * Turns Flow's company directory on or off for the whole business. It is off
 * until an owner opts in, because it shows contact details outside HR.
 */
const CompanyDirectorySetting: React.FC = () => {
  const { t } = useTranslation('hrCommon');
  const config = useGetHRConfig();
  const upsert = useUpsertHRConfig();

  if (config.isLoading) return <Skeleton height={88} />;
  if (config.isError) {
    return <QueryError onRetry={() => void config.refetch()} isRetrying={config.isFetching} />;
  }

  const stored = config.data?.directoryEnabled ?? false;
  const enabled = upsert.isPending ? (upsert.variables?.directoryEnabled ?? stored) : stored;

  return (
    <Card>
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 16 }}>
        <div style={{ flex: 1 }}>
          <div style={{ fontWeight: 600, color: 'var(--hr-text)' }}>
            {t('hrSetup.hrSettings.directoryTitle')}
          </div>
          <p style={{ marginTop: 4, fontSize: 12, color: 'var(--hr-text-muted)' }}>
            {t('hrSetup.hrSettings.directoryDescription')}
          </p>
        </div>
        <Switch
          id="hr-directory-enabled"
          aria-label={t('hrSetup.hrSettings.directoryTitle')}
          checked={enabled}
          disabled={upsert.isPending}
          onCheckedChange={(next) => upsert.mutate({ directoryEnabled: next })}
        />
      </div>
    </Card>
  );
};

export default CompanyDirectorySetting;
