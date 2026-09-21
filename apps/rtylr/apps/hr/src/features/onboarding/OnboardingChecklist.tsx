import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Rocket, CheckCircle, Upload } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useGetEmployeeChecklist,
  useListChecklistTemplates,
  useCompleteChecklistItem,
} from '../../shared/hooks/queries/useOnboarding';
import type { EmployeeChecklistItemPayload } from '@voxire/voxie/hr/onboarding/v1/types';
import {
  SectionHeader,
  Card,
  Skeleton,
  Pill,
  Button,
  Modal,
  ACCENT,
} from '../../shared/components/ui';
import DocumentUpload from '../../shared/components/DocumentUpload';

const STATUS_COLORS: Record<string, string> = {
  incomplete: 'var(--hr-status-pending)',
  in_progress: 'var(--hr-shift-blue)',
  complete: 'var(--hr-onboarding-complete)',
};

const pillStyle = (cssVar: string): React.CSSProperties => ({
  background: `color-mix(in srgb, ${cssVar} 13%, transparent)`,
  color: cssVar,
  border: `1px solid color-mix(in srgb, ${cssVar} 20%, transparent)`,
});

interface Props {
  employeeId: number;
}

const OnboardingChecklist = ({ employeeId }: Props): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading, refetch } = useGetEmployeeChecklist(employeeId);
  const { data: templatesData } = useListChecklistTemplates();
  const complete = useCompleteChecklistItem();
  const [attachmentItem, setAttachmentItem] = useState<EmployeeChecklistItemPayload | null>(null);

  const checklists = data?.checklists ?? [];
  const checklist = checklists[0];
  const items = checklist?.items ?? [];
  const templateItems =
    (templatesData?.templates ?? []).find((template) => template.id === checklist?.templateId)
      ?.items ?? [];

  const requiresAttachment = (item: EmployeeChecklistItemPayload): boolean => {
    const templateItem = templateItems.find((i) => i.id === item.checklistItemId);
    // Without the template (its query failed or the checklist was deleted),
    // keep the attachment step rather than completing an item that may need a
    // file.
    return templateItem?.requiresAttachment ?? true;
  };

  const completedCount = items.filter((i) => i.status === 'complete').length;
  const totalCount = items.length;
  const progressPct = totalCount > 0 ? Math.round((completedCount / totalCount) * 100) : 0;

  const isOverdue = (item: EmployeeChecklistItemPayload): boolean => {
    return item.status !== 'complete' && new Date(item.dueDate ?? 0).getTime() < Date.now();
  };

  const handleComplete = (item: EmployeeChecklistItemPayload): void => {
    complete.mutate({ itemId: Number(item.id ?? 0) });
  };

  const handleAttachmentUpload = (file: { name: string; url: string }): void => {
    if (!attachmentItem) return;
    complete.mutate(
      { itemId: Number(attachmentItem.id ?? 0), attachmentUrl: file.url },
      { onSuccess: () => setAttachmentItem(null) }
    );
  };

  const grouped = items.reduce<Record<string, EmployeeChecklistItemPayload[]>>((acc, item) => {
    const key = item.status ?? '';
    if (!acc[key]) acc[key] = [];
    acc[key].push(item);
    return acc;
  }, {});

  const statusOrder = ['incomplete', 'in_progress', 'complete'];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader
        title={t('onboardingMgmt.checklist.title')}
        subtitle={t('onboardingMgmt.checklist.subtitle')}
      />

      {isLoading ? (
        <Skeleton height={200} />
      ) : !checklist ? (
        <EmptyState
          icon={Rocket}
          title={t('onboardingMgmt.checklist.emptyTitle')}
          description={t('onboardingMgmt.checklist.emptyDescription')}
          action={{ label: t('common.refresh'), onClick: () => refetch() }}
        />
      ) : (
        <>
          <Card>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 8,
              }}
            >
              <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
                {t('onboardingMgmt.checklist.progress')}
              </span>
              <span style={{ fontSize: 18, fontWeight: 700, color: ACCENT }}>{progressPct}%</span>
            </div>
            <div
              style={{
                height: 6,
                borderRadius: 999,
                background: 'var(--hr-hover)',
              }}
            >
              <div
                style={{
                  width: `${progressPct}%`,
                  height: '100%',
                  background: ACCENT,
                  borderRadius: 999,
                  transition: 'width 0.3s ease',
                }}
              />
            </div>
            <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 6 }}>
              {t('onboardingMgmt.checklist.completedOf', {
                completed: completedCount,
                total: totalCount,
              })}
            </div>
          </Card>

          {statusOrder.map((status) => {
            const group = grouped[status];
            if (!group || group.length === 0) return null;
            return (
              <div key={status} style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                <h3
                  style={{
                    margin: 0,
                    fontSize: 11,
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    color: 'var(--hr-text-muted)',
                  }}
                >
                  {t(`onboardingMgmt.checklist.statuses.${status}` as const)}
                </h3>
                {group.map((item) => (
                  <Card
                    key={item.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 12,
                      flexWrap: 'wrap',
                      borderColor: isOverdue(item)
                        ? 'color-mix(in srgb, var(--hr-status-overdue) 40%, transparent)'
                        : undefined,
                    }}
                  >
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div
                        style={{
                          fontSize: 13,
                          fontWeight: 600,
                          color: 'var(--hr-text)',
                          textDecoration: item.status === 'complete' ? 'line-through' : 'none',
                          opacity: item.status === 'complete' ? 0.6 : 1,
                        }}
                      >
                        {item.checklistItemId}
                      </div>
                      <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                        {t('onboardingMgmt.checklist.dueDate', { date: item.dueDate })}
                      </div>
                    </div>
                    <Pill
                      style={pillStyle(
                        STATUS_COLORS[item.status ?? ''] ?? 'var(--hr-status-draft)'
                      )}
                    >
                      {t(`onboardingMgmt.checklist.statuses.${item.status}` as const)}
                    </Pill>
                    {isOverdue(item) && (
                      <Pill style={pillStyle('var(--hr-status-overdue)')}>
                        {t('onboardingMgmt.checklist.overdue')}
                      </Pill>
                    )}
                    {item.status !== 'complete' && (
                      <Button
                        variant="primary"
                        icon={<CheckCircle size={13} />}
                        onClick={() => {
                          if (requiresAttachment(item)) {
                            setAttachmentItem(item);
                          } else {
                            handleComplete(item);
                          }
                        }}
                        style={{ minHeight: 30, padding: '4px 10px', fontSize: 11 }}
                      >
                        {t('common.complete')}
                      </Button>
                    )}
                  </Card>
                ))}
              </div>
            );
          })}
        </>
      )}

      <Modal
        open={attachmentItem !== null}
        onClose={() => setAttachmentItem(null)}
        title={t('onboardingMgmt.checklist.completeItem')}
        width={440}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <p style={{ margin: 0, fontSize: 13, color: 'var(--hr-text-muted)' }}>
            {t('onboardingMgmt.checklist.attachmentRequired')}
          </p>
          <DocumentUpload onFile={handleAttachmentUpload} />
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            <Button variant="ghost" onClick={() => setAttachmentItem(null)}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              icon={<Upload size={13} />}
              onClick={() => {
                if (attachmentItem) handleComplete(attachmentItem);
                setAttachmentItem(null);
              }}
            >
              {t('onboardingMgmt.checklist.completeWithout')}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default OnboardingChecklist;
