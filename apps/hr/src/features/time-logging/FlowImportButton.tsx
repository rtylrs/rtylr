import React from 'react';
import { useTranslation } from 'react-i18next';
import { ArrowDownToLine } from 'lucide-react';
import { Button } from '../../shared/components/ui';

const FlowImportButton = (): React.JSX.Element => {
  const { t } = useTranslation();

  const handleClick = (): void => {
    alert(t('timeLogging.flowImportComingSoon'));
  };

  return (
    <Button variant="ghost" onClick={handleClick} icon={<ArrowDownToLine size={14} />}>
      {t('timeLogging.importFromFlow')}
    </Button>
  );
};

export default FlowImportButton;
