import React from 'react';
import { useTranslation } from 'react-i18next';
import { ArrowDownToLine } from 'lucide-react';
import { useImportFromFlow } from '../../shared/hooks/queries/useTimeEntries';
import { useHRConfigStore } from '../../shared/stores/hrConfigStore';
import { Button } from '../../shared/components/ui';

const FlowImportButton = (): React.JSX.Element | null => {
  const { t } = useTranslation();
  const config = useHRConfigStore((s) => s.config);
  const importFromFlow = useImportFromFlow();

  const isFlowSubscribed = config?.subscribedApps?.includes('flow');

  if (!isFlowSubscribed) {
    return null;
  }

  const handleClick = (): void => {
    importFromFlow.mutate();
  };

  return (
    <Button
      variant="ghost"
      onClick={handleClick}
      disabled={importFromFlow.isPending}
      icon={<ArrowDownToLine size={14} />}
    >
      {importFromFlow.isPending
        ? t('timeLogging.flowImport.importing')
        : t('timeLogging.importFromFlow')}
    </Button>
  );
};

export default FlowImportButton;
