import React from 'react';
import { useTranslation } from 'react-i18next';
import { HardDrive } from 'lucide-react';
import { useListAssets } from '../../shared/hooks/queries/useEquipment';
import { SectionHeader, Card, EmptyState, Skeleton, Pill } from '../../shared/components/ui';

const CONDITION_COLORS: Record<string, string> = {
  new: '#10b981',
  good: '#3b82f6',
  fair: '#f59e0b',
  damaged: '#ef4444',
};

const AssetRegister = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading } = useListAssets();
  const assets = data?.assets ?? [];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader
        title={t('equipment.assets.title')}
        subtitle={t('equipment.assets.subtitle')}
      />

      {isLoading ? (
        <Skeleton height={300} />
      ) : assets.length === 0 ? (
        <EmptyState
          icon={<HardDrive size={20} />}
          title={t('equipment.assets.emptyTitle')}
          description={t('equipment.assets.emptyDescription')}
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
                    t('equipment.assets.table.name'),
                    t('equipment.assets.table.serialNumber'),
                    t('equipment.assets.table.assignedDate'),
                    t('equipment.assets.table.condition'),
                    t('equipment.assets.table.returnedDate'),
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
                    <td style={{ padding: '10px 12px', color: 'var(--hr-text)', fontWeight: 600 }}>
                      {asset.name}
                    </td>
                    <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                      {asset.serialNumber ?? '—'}
                    </td>
                    <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                      {asset.assignedDate}
                    </td>
                    <td style={{ padding: '10px 12px' }}>
                      <Pill color={CONDITION_COLORS[asset.condition] ?? '#a1a1aa'}>
                        {t(`equipment.assets.conditions.${asset.condition}` as const)}
                      </Pill>
                    </td>
                    <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                      {asset.returnedDate ?? '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}
    </div>
  );
};

export default AssetRegister;
