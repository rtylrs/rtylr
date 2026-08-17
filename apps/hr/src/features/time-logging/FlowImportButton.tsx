import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ArrowDownToLine } from 'lucide-react';
import { useImportFromFlow } from '../../shared/hooks/queries/useTimeEntries';
import { useMyEmployeeId } from '../../shared/hooks/useHrExtras';
import { useHRConfigStore } from '../../shared/stores/hrConfigStore';
import { getWeekRange } from '../../shared/utils/dateRange';
import { flowImportFormSchema } from '../../shared/schemas/timeEntries';
import { Button, Modal, Label } from '../../shared/components/ui';
import DateRangeFilter from '../../shared/components/DateRangeFilter';

const FlowImportButton = (): React.JSX.Element | null => {
  const { t } = useTranslation();
  const config = useHRConfigStore((s) => s.config);
  const { data: myEmployeeId, isError: employeeLookupFailed } = useMyEmployeeId();
  const importFromFlow = useImportFromFlow();

  const [modalOpen, setModalOpen] = useState(false);
  const [range, setRange] = useState(getWeekRange);
  const [error, setError] = useState<string | undefined>();

  const isFlowSubscribed = config?.subscribedApps?.includes('flow');

  if (!isFlowSubscribed) {
    return null;
  }

  const handleOpen = (): void => {
    setRange(getWeekRange());
    setError(undefined);
    setModalOpen(true);
  };

  const handleClose = (): void => {
    setModalOpen(false);
  };

  const handleRangeChange = (from: string, to: string): void => {
    setRange({ start: from, end: to });
    setError(undefined);
  };

  const handleConfirm = (): void => {
    const result = flowImportFormSchema.safeParse({
      employeeId: myEmployeeId ?? 0,
      periodStart: range.start,
      periodEnd: range.end,
    });

    if (!result.success) {
      setError(t('timeLogging.flowImport.invalidRange'));
      return;
    }

    setError(undefined);
    importFromFlow.mutate(result.data, {
      onSuccess: () => {
        setModalOpen(false);
      },
    });
  };

  return (
    <>
      <Button
        variant="ghost"
        onClick={handleOpen}
        disabled={importFromFlow.isPending}
        icon={<ArrowDownToLine size={14} />}
      >
        {t('timeLogging.importFromFlow')}
      </Button>

      <Modal
        open={modalOpen}
        onClose={handleClose}
        title={t('timeLogging.flowImport.modalTitle')}
        width={380}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div>
            <Label>{t('timeLogging.flowImport.periodLabel')}</Label>
            <DateRangeFilter from={range.start} to={range.end} onChange={handleRangeChange} />
          </div>

          {error && <p style={{ color: 'var(--hr-error)', fontSize: 12, margin: 0 }}>{error}</p>}
          {employeeLookupFailed && (
            <p style={{ color: 'var(--hr-error)', fontSize: 12, margin: 0 }}>
              {t('timeLogging.form.employeeLookupFailed')}
            </p>
          )}

          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 4 }}>
            <Button variant="ghost" type="button" onClick={handleClose}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              type="button"
              onClick={handleConfirm}
              disabled={importFromFlow.isPending || myEmployeeId === undefined}
            >
              {importFromFlow.isPending
                ? t('timeLogging.flowImport.importing')
                : t('timeLogging.flowImport.confirm')}
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
};

export default FlowImportButton;
