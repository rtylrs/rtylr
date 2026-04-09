import React, { useState } from 'react';
import { toast } from 'sonner';
import { FileText, Trash2, AlertTriangle } from 'lucide-react';
import { useDocuments, useAddDocument, useDeleteDocument, useExpiringDocuments } from '../../hooks/useHrExtras';
import { useStaff } from '../../hooks/useStaff';
import { Card, SectionHeader, EmptyState, Skeleton, Pill, Select, Label, ConfirmDialog } from '../../components/ui';
import DocumentUpload from '../../components/DocumentUpload';

const TYPES = ['I-9', 'W4', 'Contract', 'NDA', 'Offer Letter', 'Other'] as const;

const Documents: React.FC = () => {
  const { data, isLoading } = useDocuments();
  const { data: expiring = [] } = useExpiringDocuments(30);
  const { data: staff = [] } = useStaff();
  const add = useAddDocument();
  const del = useDeleteDocument();

  const [staffId, setStaffId] = useState(0);
  const [type, setType] = useState<(typeof TYPES)[number]>('Other');
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const onFile = (f: { name: string; sizeKb: number }) => {
    if (!staffId) {
      toast.error('Select an employee first');
      return;
    }
    const who = staff.find((s) => s.id === staffId);
    if (!who) return;
    add.mutate(
      { staffId: who.id, staffName: who.name, name: f.name, type, sizeKb: f.sizeKb },
      { onSuccess: () => toast.success('Document uploaded') },
    );
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader title="Documents" subtitle="Employee document vault" />

      {expiring.length > 0 && (
        <Card style={{ display: 'flex', alignItems: 'center', gap: 12, borderColor: 'rgba(245,158,11,0.4)', background: 'rgba(245,158,11,0.06)' }}>
          <span style={{ padding: 8, borderRadius: 8, background: 'rgba(245,158,11,0.15)', color: '#f59e0b' }}>
            <AlertTriangle size={16} />
          </span>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>{expiring.length} document(s) expiring soon</div>
            <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
              {expiring.map((d) => `${d.name} (${d.expiresAt})`).join(', ')}
            </div>
          </div>
        </Card>
      )}

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        <div>
          <Label>Employee</Label>
          <Select value={staffId} onChange={(e) => setStaffId(Number(e.target.value))}>
            <option value={0}>Select</option>
            {staff.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
          </Select>
        </div>
        <div>
          <Label>Type</Label>
          <Select value={type} onChange={(e) => setType(e.target.value as (typeof TYPES)[number])}>
            {TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
          </Select>
        </div>
      </div>

      <DocumentUpload onFile={onFile} />

      {isLoading ? (
        <Skeleton height={200} />
      ) : !data || data.length === 0 ? (
        <EmptyState icon={<FileText size={20} />} title="No documents yet" description="Upload I-9, W4, contracts, and other files." />
      ) : (
        <Card style={{ padding: 0 }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
            <thead>
              <tr style={{ background: 'var(--hr-hover)' }}>
                <th style={{ textAlign: 'left', padding: '10px 12px', fontSize: 10, color: 'var(--hr-text-muted)', textTransform: 'uppercase', fontWeight: 600 }}>Document</th>
                <th style={{ textAlign: 'left', padding: '10px 12px', fontSize: 10, color: 'var(--hr-text-muted)', textTransform: 'uppercase', fontWeight: 600 }}>Employee</th>
                <th style={{ textAlign: 'left', padding: '10px 12px', fontSize: 10, color: 'var(--hr-text-muted)', textTransform: 'uppercase', fontWeight: 600 }}>Type</th>
                <th style={{ textAlign: 'left', padding: '10px 12px', fontSize: 10, color: 'var(--hr-text-muted)', textTransform: 'uppercase', fontWeight: 600 }}>Uploaded</th>
                <th style={{ textAlign: 'right', padding: '10px 12px' }}></th>
              </tr>
            </thead>
            <tbody>
              {data.map((d) => (
                <tr key={d.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                  <td style={{ padding: '10px 12px', color: 'var(--hr-text)', fontWeight: 600 }}>{d.name}</td>
                  <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>{d.staffName}</td>
                  <td style={{ padding: '10px 12px' }}><Pill>{d.type}</Pill></td>
                  <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>{d.uploadedAt}</td>
                  <td style={{ padding: '10px 12px', textAlign: 'right' }}>
                    <button
                      onClick={() => setDeleteId(d.id)}
                      style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#ef4444' }}
                    >
                      <Trash2 size={13} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}

      <ConfirmDialog
        open={deleteId !== null}
        title="Delete document?"
        description="This action cannot be undone."
        variant="danger"
        confirmLabel="Delete"
        onCancel={() => setDeleteId(null)}
        onConfirm={() => {
          if (deleteId !== null) {
            del.mutate(deleteId, { onSuccess: () => toast.success('Document deleted') });
            setDeleteId(null);
          }
        }}
      />
    </div>
  );
};

export default Documents;
