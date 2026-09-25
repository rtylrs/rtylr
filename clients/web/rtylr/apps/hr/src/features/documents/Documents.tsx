import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { FileText, Trash2, AlertTriangle } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useDocuments,
  useAddDocument,
  useDeleteDocument,
  useExpiringDocuments,
} from '../../shared/hooks/useHrExtras';
import { useStaff } from '../staff/useStaff';
import {
  Card,
  SectionHeader,
  Skeleton,
  Pill,
  Select,
  Label,
  ConfirmDialog,
  useIsMobile,
} from '../../shared/components/ui';
import DocumentUpload from '../../shared/components/DocumentUpload';

const TYPES = ['I-9', 'W4', 'Contract', 'NDA', 'Offer Letter', 'Other'] as const;

const Documents: React.FC = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile(900);
  const { data, isLoading, refetch } = useDocuments();
  const { data: expiring = [] } = useExpiringDocuments(30);
  const { data: staff = [] } = useStaff();
  const add = useAddDocument();
  const del = useDeleteDocument();

  const [staffId, setStaffId] = useState(0);
  const [type, setType] = useState<(typeof TYPES)[number]>('Other');
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const onFile = (f: { name: string; sizeKb: number; url: string }): void => {
    if (!staffId) {
      toast.error(t('documents.selectEmployeeFirst'));
      return;
    }
    const who = staff.find((s) => s.id === staffId);
    if (!who) return;
    add.mutate(
      { staffId: who.id, staffName: who.name, name: f.name, type, sizeKb: f.sizeKb, url: f.url },
      { onSuccess: () => toast.success(t('documents.uploaded')) }
    );
  };

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.documents'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader title={t('documents.title')} subtitle={t('documents.subtitle')} />

        {expiring.length > 0 && (
          <Card
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              borderColor: 'color-mix(in srgb, var(--hr-status-pending) 40%, transparent)',
              background: 'color-mix(in srgb, var(--hr-status-pending) 6%, transparent)',
            }}
          >
            <span
              style={{
                padding: 8,
                borderRadius: 8,
                background: 'color-mix(in srgb, var(--hr-status-pending) 15%, transparent)',
                color: 'var(--hr-status-pending)',
              }}
            >
              <AlertTriangle size={16} />
            </span>
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
                {t('documents.expiringSoon', { count: expiring.length })}
              </div>
              <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                {expiring.map((d) => `${d.name} (${d.expiresAt})`).join(', ')}
              </div>
            </div>
          </Card>
        )}

        <div
          style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: 16 }}
        >
          <div>
            <Label>{t('documents.filterEmployee')}</Label>
            <Select value={staffId} onChange={(e) => setStaffId(Number(e.target.value))}>
              <option value={0}>{t('documents.filterSelect')}</option>
              {staff.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <Label>{t('documents.filterType')}</Label>
            <Select
              value={type}
              onChange={(e) => setType(e.target.value as (typeof TYPES)[number])}
            >
              {TYPES.map((docType) => (
                <option key={docType} value={docType}>
                  {docType}
                </option>
              ))}
            </Select>
          </div>
        </div>

        <DocumentUpload onFile={onFile} />

        {isLoading ? (
          <Skeleton height={200} />
        ) : !data || data.length === 0 ? (
          <EmptyState
            icon={FileText}
            title={t('documents.emptyTitle')}
            description={t('documents.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => refetch() }}
          />
        ) : (
          <Card style={{ padding: 0, overflow: 'hidden' }}>
            <div style={{ overflowX: 'auto' }}>
              <table
                style={{ width: '100%', minWidth: 640, borderCollapse: 'collapse', fontSize: 12 }}
              >
                <thead>
                  <tr style={{ background: 'var(--hr-hover)' }}>
                    {[
                      t('documents.table.document'),
                      t('documents.table.employee'),
                      t('documents.table.type'),
                      t('documents.table.uploaded'),
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
                    <th style={{ textAlign: 'right', padding: '10px 12px' }}></th>
                  </tr>
                </thead>
                <tbody>
                  {data.map((d) => (
                    <tr key={d.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                      <td
                        style={{ padding: '10px 12px', color: 'var(--hr-text)', fontWeight: 600 }}
                      >
                        {d.name}
                      </td>
                      <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                        {d.staffName}
                      </td>
                      <td style={{ padding: '10px 12px' }}>
                        <Pill>{d.type}</Pill>
                      </td>
                      <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                        {d.uploadedAt}
                      </td>
                      <td style={{ padding: '10px 12px', textAlign: 'right' }}>
                        <button
                          type="button"
                          onClick={() => setDeleteId(d.id)}
                          style={{
                            background: 'transparent',
                            border: 'none',
                            cursor: 'pointer',
                            color: 'var(--hr-error)',
                          }}
                          aria-label={t('common.delete')}
                        >
                          <Trash2 size={13} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>
        )}

        <ConfirmDialog
          open={deleteId !== null}
          title={t('documents.deleteTitle')}
          description={t('documents.deleteConfirm')}
          variant="danger"
          confirmLabel={t('documents.deleteButton')}
          onCancel={() => setDeleteId(null)}
          onConfirm={() => {
            if (deleteId !== null) {
              del.mutate(deleteId, { onSuccess: () => toast.success(t('documents.deleted')) });
              setDeleteId(null);
            }
          }}
        />
      </div>
    </>
  );
};

export default Documents;
