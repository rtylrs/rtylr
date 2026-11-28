import React from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Check, FileText } from 'lucide-react';
import { Badge, Button, Select } from '../../shared/components/ui';
import PagesEditor from '../design/PagesEditor';
import { useMenuBranchProfile } from './useMenuBranchProfile';

const MenuPagesPage: React.FC = () => {
  const { t } = useTranslation();
  const {
    user,
    branches,
    selectedLocationId,
    setSelectedLocationId,
    selectedBranch,
    remoteProfile,
    draftConfig,
    patchConfig,
    saved,
    isSaving,
    isPublishing,
    saveDraft,
    publishLive,
  } = useMenuBranchProfile();

  const u = user as Record<string, unknown> | null;
  const defaultAddress = selectedBranch?.address || (u?.address as string) || '';
  const defaultPhone = (u?.phone as string) || '';

  return (
    <>
      <HelmetContainer title={`${t('common.nav.menuPages')} · rtylr`} />
      <div className="w-full max-w-4xl space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <FileText size={18} style={{ color: 'var(--dash-text-muted)' }} />
              <h1 className="text-lg font-semibold">{t('common.nav.menuPages')}</h1>
              {remoteProfile?.publishedAt && (
                <Badge variant="success">{t('design.studio.liveBadge')}</Badge>
              )}
            </div>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              {t('design.pagesEditor.subtitle')}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {branches.length > 0 && (
              <Select
                aria-label={t('design.studio.branchLabel')}
                value={String(selectedLocationId)}
                onChange={(e) => setSelectedLocationId(Number(e.target.value) || 0)}
                className="h-8 text-xs w-44"
              >
                <option value="0">{t('design.studio.defaultBranch')}</option>
                {branches.map((branch) => (
                  <option key={branch.id} value={String(branch.id)}>
                    {branch.name}
                  </option>
                ))}
              </Select>
            )}
            <Button
              size="sm"
              variant="outline"
              disabled={isSaving}
              onClick={() => void saveDraft()}
            >
              {saved ? (
                <>
                  <Check className="w-3.5 h-3.5 me-1" />
                  {t('design.studio.saved')}
                </>
              ) : (
                t('design.studio.save')
              )}
            </Button>
            <Button
              size="sm"
              variant="accent"
              disabled={isSaving || isPublishing}
              onClick={() => void publishLive()}
            >
              {t('design.studio.publish')}
            </Button>
          </div>
        </div>

        <div
          className="rounded-xl border p-5"
          style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
        >
          <PagesEditor
            selectedLocationId={selectedLocationId}
            pages={draftConfig.pages}
            content={draftConfig.content ?? {}}
            defaultAddress={defaultAddress}
            defaultPhone={defaultPhone}
            hideReviews
            onPatchPages={(patch) => patchConfig({ pages: { ...draftConfig.pages, ...patch } })}
            onPatchContent={(patch) =>
              patchConfig({ content: { ...draftConfig.content, ...patch } })
            }
          />
        </div>
      </div>
    </>
  );
};

export default MenuPagesPage;
