import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { FileText, Upload, AlertTriangle } from 'lucide-react';
import {
  useListEmployeeDocuments,
  useUploadDocument,
} from '../../shared/hooks/queries/useDocuments';
import type { EmployeeDocumentPayload } from '@voxire/voxie/hr/document/v1/types';
import {
  Card,
  EmptyState,
  Skeleton,
  Pill,
  Button,
  Modal,
  Input,
  Select,
  Label,
} from '../../shared/components/ui';
import DocumentUpload from '../../shared/components/DocumentUpload';

const DOCUMENT_TYPES = ['id_card', 'contract', 'certificate', 'visa', 'other'] as const;

const EXPIRY_WARNING_DAYS = 14;

function isExpiringSoon(expiryDate: string | undefined): boolean {
  if (!expiryDate) return false;
  const diff = new Date(expiryDate).getTime() - Date.now();
  return diff > 0 && diff < EXPIRY_WARNING_DAYS * 24 * 60 * 60 * 1000;
}

function isExpired(expiryDate: string | undefined): boolean {
  if (!expiryDate) return false;
  return new Date(expiryDate).getTime() < Date.now();
}

interface Props {
  employeeId: number;
}

const DocumentVault = ({ employeeId }: Props): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading } = useListEmployeeDocuments(employeeId);
  const upload = useUploadDocument();
  const [showUpload, setShowUpload] = useState(false);
  const [docName, setDocName] = useState('');
  const [docType, setDocType] = useState<string>('other');
  const [expiryDate, setExpiryDate] = useState('');

  const documents = data?.documents ?? [];
  const expiringDocs = documents.filter(
    (d) => isExpiringSoon(d.expiryDate) || isExpired(d.expiryDate)
  );

  const handleFile = (file: { name: string; sizeKb: number; url: string }): void => {
    upload.mutate(
      {
        employeeId,
        documentType: docType,
        title: docName || file.name,
        storageMode: 'stored',
        fileUrl: file.url,
        expiryDate: expiryDate || undefined,
      },
      {
        onSuccess: () => {
          setShowUpload(false);
          setDocName('');
          setDocType('other');
          setExpiryDate('');
        },
      }
    );
  };

  const expiryStyle = (doc: EmployeeDocumentPayload): React.CSSProperties => {
    if (isExpired(doc.expiryDate)) return { color: 'var(--hr-error)', fontWeight: 600 };
    if (isExpiringSoon(doc.expiryDate))
      return { color: 'var(--hr-status-pending)', fontWeight: 600 };
    return { color: 'var(--hr-text-muted)' };
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 700, color: 'var(--hr-text)' }}>
          {t('documents.vault.title')}
        </h3>
        <Button
          variant="primary"
          icon={<Upload size={13} />}
          onClick={() => setShowUpload(true)}
          style={{ minHeight: 30, padding: '4px 10px', fontSize: 11 }}
        >
          {t('documents.vault.upload')}
        </Button>
      </div>

      {expiringDocs.length > 0 && (
        <Card
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            borderColor: 'color-mix(in srgb, var(--hr-status-pending) 40%, transparent)',
            background: 'color-mix(in srgb, var(--hr-status-pending) 6%, transparent)',
            padding: 12,
          }}
        >
          <span
            style={{
              padding: 6,
              borderRadius: 8,
              background: 'color-mix(in srgb, var(--hr-status-pending) 15%, transparent)',
              color: 'var(--hr-status-pending)',
            }}
          >
            <AlertTriangle size={14} />
          </span>
          <span style={{ fontSize: 12, color: 'var(--hr-text)' }}>
            {t('documents.vault.expiryWarning', { count: expiringDocs.length })}
          </span>
        </Card>
      )}

      {isLoading ? (
        <Skeleton height={200} />
      ) : documents.length === 0 ? (
        <EmptyState
          icon={<FileText size={20} />}
          title={t('documents.vault.emptyTitle')}
          description={t('documents.vault.emptyDescription')}
          action={
            <Button
              variant="primary"
              icon={<Upload size={13} />}
              onClick={() => setShowUpload(true)}
            >
              {t('documents.vault.upload')}
            </Button>
          }
        />
      ) : (
        <Card style={{ padding: 0, overflow: 'hidden' }}>
          <div style={{ overflowX: 'auto' }}>
            <table
              style={{ width: '100%', minWidth: 560, borderCollapse: 'collapse', fontSize: 12 }}
            >
              <thead>
                <tr style={{ background: 'var(--hr-hover)' }}>
                  {[
                    t('documents.vault.table.name'),
                    t('documents.vault.table.type'),
                    t('documents.vault.table.storageMode'),
                    t('documents.vault.table.expiryDate'),
                  ].map((h) => (
                    <th
                      key={h}
                      style={{
                        textAlign: 'left',
                        padding: '10px 12px',
                        fontSize: 10,
                        color: 'var(--hr-text-muted)',
                        textTransform: 'uppercase',
                        fontWeight: 600,
                      }}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {documents.map((doc) => (
                  <tr key={doc.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                    <td style={{ padding: '10px 12px', color: 'var(--hr-text)', fontWeight: 600 }}>
                      {doc.name}
                    </td>
                    <td style={{ padding: '10px 12px' }}>
                      <Pill>{t(`documents.vault.types.${doc.documentType}` as const)}</Pill>
                    </td>
                    <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                      {t(`documents.vault.storageModes.${doc.storageMode}` as const)}
                    </td>
                    <td style={{ padding: '10px 12px', ...expiryStyle(doc) }}>
                      {doc.expiryDate ?? '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      <Modal
        open={showUpload}
        onClose={() => setShowUpload(false)}
        title={t('documents.vault.uploadTitle')}
        width={440}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div>
            <Label>{t('documents.vault.form.name')}</Label>
            <Input
              value={docName}
              onChange={(e) => setDocName(e.target.value)}
              placeholder={t('documents.vault.form.namePlaceholder')}
              style={{ fontSize: 16 }}
            />
          </div>
          <div>
            <Label>{t('documents.vault.form.type')}</Label>
            <Select
              value={docType}
              onChange={(e) => setDocType(e.target.value)}
              style={{ fontSize: 16 }}
            >
              {DOCUMENT_TYPES.map((dt) => (
                <option key={dt} value={dt}>
                  {t(`documents.vault.types.${dt}` as const)}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <Label>{t('documents.vault.form.expiryDate')}</Label>
            <Input
              type="date"
              value={expiryDate}
              onChange={(e) => setExpiryDate(e.target.value)}
              style={{ fontSize: 16 }}
            />
          </div>
          <DocumentUpload onFile={handleFile} />
        </div>
      </Modal>
    </div>
  );
};

export default DocumentVault;
