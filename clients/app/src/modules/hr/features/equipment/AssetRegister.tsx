import React from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { HardDrive } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useListAssets } from '../../shared/hooks/queries/useEquipment';
import { SectionHeader, Card, Skeleton, Pill } from '../../shared/components/ui';

const pillStyle = (cssVar: string): React.CSSProperties => ({
  background: `color-mix(in srgb, ${cssVar} 13%, transparent)`,
  color: cssVar,
  border: `1px solid color-mix(in srgb, ${cssVar} 20%, transparent)`,
});

const CONDITION_COLORS: Record<string, string> = {
  new: 'var(--hr-status-approved)',
  good: 'var(--hr-shift-blue)',
  fair: 'var(--hr-status-pending)',
  damaged: 'var(--hr-error)',
};

const AssetRegister = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const { data, isLoading, refetch } = useListAssets();
  const assets = data?.assets ?? [];

  return (
    <>
      <HelmetContainer title={pageTitle(t('hrNav.assets'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('hrEquipment.assets.title')}
          subtitle={t('hrEquipment.assets.subtitle')}
        />

        {isLoading ? (
          <Skeleton height={300} />
        ) : assets.length === 0 ? (
          <EmptyState
            icon={HardDrive}
            title={t('hrEquipment.assets.emptyTitle')}
            description={t('hrEquipment.assets.emptyDescription')}
            action={{ label: t('hrCommon.refresh'), onClick: () => refetch() }}
          />
        ) : (
          <Card style={{ padding: 0, overflow: 'hidden' }}>
            <div style={{ overflowX: 'auto' }}>
              <table
                style={{ width: '100%', minWidth: 700, borderCollapse: 'collapse', fontSize: 12 }}
              >
                <thead>
                  <tr style={{ background: 'var(--hr-hover)' }}>
                    {[
                      t('hrEquipment.assets.table.name'),
                      t('hrEquipment.assets.table.serialNumber'),
                      t('hrEquipment.assets.table.assignedDate'),
                      t('hrEquipment.assets.table.condition'),
                      t('hrEquipment.assets.table.returnedDate'),
                    ].map((header) => (
                      <th
                        key={header}
                        style={{
                          textAlign: 'left',
                          padding: '10px 12px',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                          fontWeight: 600,
                        }}
                      >
                        {header}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {assets.map((asset) => (
                    <tr key={asset.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                      <td
                        style={{ padding: '10px 12px', color: 'var(--hr-text)', fontWeight: 600 }}
                      >
                        {asset.name}
                      </td>
                      <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                        {asset.serialNumber ?? '-'}
                      </td>
                      <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                        {asset.assignedDate}
                      </td>
                      <td style={{ padding: '10px 12px' }}>
                        <Pill
                          style={pillStyle(
                            CONDITION_COLORS[asset.condition ?? ''] ?? 'var(--hr-status-draft)'
                          )}
                        >
                          {t(`hrEquipment.assets.conditions.${asset.condition}` as const)}
                        </Pill>
                      </td>
                      <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                        {asset.returnedDate ?? '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>
        )}
      </div>
    </>
  );
};

export default AssetRegister;
