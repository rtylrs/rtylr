import { useTranslation } from 'react-i18next';
import { Spinner } from '@voxire/ui';
export default function AdminQueryState({
  loading = false,
  retry,
}: {
  loading?: boolean;
  retry?: () => void;
}) {
  const { t } = useTranslation('recruitCommon');
  if (loading)
    return (
      <div className="flex min-h-48 items-center justify-center">
        <Spinner />
      </div>
    );
  return (
    <div role="alert" className="flex flex-col items-center gap-4 py-12">
      <p>{t('recruitCommon.errors.load')}</p>
      <button type="button" className="rounded-lg bg-[var(--dash-hover)] px-4 py-2" onClick={retry}>
        {t('recruitCommon.retry')}
      </button>
    </div>
  );
}
