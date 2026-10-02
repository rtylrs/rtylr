import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Plus, FileText } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useListChecklistTemplates,
  useCreateChecklistTemplate,
} from '../../shared/hooks/queries/useOnboarding';
import {
  SectionHeader,
  Card,
  Skeleton,
  Pill,
  Button,
  Modal,
  Input,
  Select,
  Label,
  ACCENT,
} from '../../shared/components/ui';

const ChecklistTemplates = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const { data, isLoading } = useListChecklistTemplates();
  const createTemplate = useCreateChecklistTemplate();
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState('');
  const [triggerType, setTriggerType] = useState<string>('onboarding');

  const templates = data?.templates ?? [];

  const handleCreate = (): void => {
    if (!name.trim()) {
      toast.error(t('hrOnboarding.templateNameRequired'));
      return;
    }
    createTemplate.mutate(
      { name, triggerType },
      {
        onSuccess: () => {
          setShowForm(false);
          setName('');
          setTriggerType('onboarding');
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title={pageTitle(t('hrCommon.pageTitles.checklistTemplates'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('hrOnboardingMgmt.templates.title')}
          subtitle={t('hrOnboardingMgmt.templates.subtitle')}
          right={
            <Button variant="primary" icon={<Plus size={14} />} onClick={() => setShowForm(true)}>
              {t('hrOnboardingMgmt.templates.create')}
            </Button>
          }
        />

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Skeleton height={80} />
            <Skeleton height={80} />
          </div>
        ) : templates.length === 0 ? (
          <EmptyState
            icon={FileText}
            title={t('hrOnboardingMgmt.templates.emptyTitle')}
            description={t('hrOnboardingMgmt.templates.emptyDescription')}
            action={{
              label: t('hrOnboardingMgmt.templates.create'),
              onClick: () => setShowForm(true),
            }}
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {templates.map((tmpl) => (
              <Card
                key={tmpl.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  flexWrap: 'wrap',
                }}
              >
                <div
                  style={{
                    width: 36,
                    height: 36,
                    borderRadius: 10,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: `${ACCENT}22`,
                    color: ACCENT,
                    flexShrink: 0,
                  }}
                >
                  <FileText size={16} />
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
                    {tmpl.name}
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                    {t('hrOnboardingMgmt.templates.itemCount', {
                      count: (tmpl.items ?? []).length,
                    })}
                  </div>
                </div>
                <Pill
                  style={
                    tmpl.triggerType === 'onboarding'
                      ? {
                          background:
                            'color-mix(in srgb, var(--hr-onboarding-complete) 13%, transparent)',
                          color: 'var(--hr-onboarding-complete)',
                          border:
                            '1px solid color-mix(in srgb, var(--hr-onboarding-complete) 20%, transparent)',
                        }
                      : {
                          background:
                            'color-mix(in srgb, var(--hr-status-pending) 13%, transparent)',
                          color: 'var(--hr-status-pending)',
                          border:
                            '1px solid color-mix(in srgb, var(--hr-status-pending) 20%, transparent)',
                        }
                  }
                >
                  {t(`hrOnboardingMgmt.templates.triggers.${tmpl.triggerType}` as const)}
                </Pill>
              </Card>
            ))}
          </div>
        )}

        <Modal
          open={showForm}
          onClose={() => setShowForm(false)}
          title={t('hrOnboardingMgmt.templates.form.title')}
          width={440}
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div>
              <Label>{t('hrOnboardingMgmt.templates.form.name')}</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={t('hrOnboardingMgmt.templates.form.namePlaceholder')}
                style={{ fontSize: 16 }}
              />
            </div>
            <div>
              <Label>{t('hrOnboardingMgmt.templates.form.triggerType')}</Label>
              <Select
                value={triggerType}
                onChange={(e) => setTriggerType(e.target.value)}
                style={{ fontSize: 16 }}
              >
                <option value="onboarding">
                  {t('hrOnboardingMgmt.templates.triggers.onboarding')}
                </option>
                <option value="offboarding">
                  {t('hrOnboardingMgmt.templates.triggers.offboarding')}
                </option>
              </Select>
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 4 }}>
              <Button variant="ghost" onClick={() => setShowForm(false)}>
                {t('hrCommon.cancel')}
              </Button>
              <Button
                variant="primary"
                onClick={handleCreate}
                disabled={createTemplate.isPending || !name.trim()}
              >
                {t('hrOnboardingMgmt.templates.form.submit')}
              </Button>
            </div>
          </div>
        </Modal>
      </div>
    </>
  );
};

export default ChecklistTemplates;
