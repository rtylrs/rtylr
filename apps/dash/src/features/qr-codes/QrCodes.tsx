import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { Plus, QrCode as QrCodeIcon, Copy, ExternalLink } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Button,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  Badge,
} from '../../shared/components/ui';
import { SkeletonList } from '../../shared/components/dash';
import { unwrap } from '@voxire/api';
import { safe } from '../../shared/lib/api';

interface TableRow {
  id: number;
  name: string;
  capacity: number;
  status: string;
  qrCodeUrl: string;
}

const useQrTables = () =>
  useQuery<TableRow[]>({
    queryKey: ['qr-tables'],
    queryFn: async () => {
      try {
        const res = await unwrap(safe.get<{ tables?: TableRow[] } | TableRow[]>('/pos/tables'));
        return Array.isArray(res) ? res : (res.tables ?? []);
      } catch {
        return [];
      }
    },
  });

const QrCodes: React.FC = () => {
  const navigate = useNavigate();
  const { data: tables = [], isLoading } = useQrTables();

  return (
    <>
      <HelmetContainer title="QR codes · rtylr" />
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
              QR Codes
            </h1>
            <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
              Manage your table QR codes
            </p>
          </div>
          <Button size="sm" onClick={() => navigate({ to: '/qr-codes/create' })}>
            <Plus size={14} />
            Generate QR Code
          </Button>
        </div>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle>All QR Codes</CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            {isLoading ? (
              <SkeletonList rows={4} withAvatar={false} />
            ) : tables.length === 0 ? (
              <Table>
                <TableBody>
                  <TableRow>
                    <TableCell colSpan={5}>
                      <div className="flex flex-col items-center justify-center py-12 text-center">
                        <div
                          className="h-16 w-16 rounded-2xl flex items-center justify-center mb-4"
                          style={{ background: 'var(--dash-hover)' }}
                        >
                          <QrCodeIcon size={28} style={{ color: 'var(--dash-text-dim)' }} />
                        </div>
                        <p
                          className="text-sm font-medium"
                          style={{ color: 'var(--dash-text-muted)' }}
                        >
                          No QR codes yet
                        </p>
                        <p className="text-xs mt-1 mb-4" style={{ color: 'var(--dash-text-dim)' }}>
                          Generate QR codes for your tables to let customers order easily
                        </p>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => navigate({ to: '/qr-codes/create' })}
                        >
                          <Plus size={13} />
                          Generate your first QR code
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Capacity</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tables.map((t) => (
                    <TableRow key={t.id}>
                      <TableCell className="font-medium">{t.name}</TableCell>
                      <TableCell>{t.capacity}</TableCell>
                      <TableCell>
                        <Badge variant={t.status === 'available' ? 'default' : 'secondary'}>
                          {t.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center gap-2 justify-end">
                          {t.qrCodeUrl && (
                            <>
                              <button
                                className="h-7 w-7 rounded-lg flex items-center justify-center transition-colors hover:bg-[var(--dash-hover)]"
                                style={{ color: 'var(--dash-text-dim)' }}
                                aria-label="Copy QR URL"
                                onClick={() => {
                                  void navigator.clipboard.writeText(t.qrCodeUrl);
                                  toast.success('URL copied');
                                }}
                              >
                                <Copy size={13} />
                              </button>
                              <a
                                href={t.qrCodeUrl}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="h-7 w-7 rounded-lg flex items-center justify-center transition-colors hover:bg-[var(--dash-hover)]"
                                style={{ color: 'var(--dash-text-dim)' }}
                                aria-label="Open QR URL"
                              >
                                <ExternalLink size={13} />
                              </a>
                            </>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>
    </>
  );
};

export default QrCodes;
