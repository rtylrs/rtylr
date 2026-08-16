import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { Copy, ExternalLink, Plus, QrCode as QrCodeIcon, Search } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Badge, Button } from '../../shared/components/ui';
import { unwrap } from '@voxire/api';
import { safe } from '../../shared/lib/api';

interface QrTable {
  id: number;
  name: string;
  capacity: number;
  status: string;
  qrCodeUrl: string;
}

const useQrTables = () =>
  useQuery<QrTable[]>({
    queryKey: ['qr-tables'],
    queryFn: async () => {
      const res = await unwrap(
        safe.get<{ tables?: QrTable[] } | QrTable[]>('/pos/table/v1/tables')
      );
      return Array.isArray(res) ? res : (res.tables ?? []);
    },
  });

function formatStatus(status: string): string {
  return status
    .split(/[_-]/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

const QrCodes: React.FC = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState('');
  const { data: tables = [], isLoading } = useQrTables();
  const normalizedSearch = searchTerm.trim().toLowerCase();

  const filteredTables = useMemo(
    () =>
      tables.filter((table) =>
        normalizedSearch ? table.name.toLowerCase().includes(normalizedSearch) : true
      ),
    [normalizedSearch, tables]
  );

  const readyTables = tables.filter((table) => Boolean(table.qrCodeUrl)).length;
  const totalCapacity = tables.reduce((sum, table) => sum + (table.capacity || 0), 0);

  return (
    <>
      <HelmetContainer title="QR codes · rtylr" />
      <div className="w-full min-w-0 space-y-5">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <QrCodeIcon size={17} style={{ color: 'var(--dash-text-muted)' }} />
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                QR Codes
              </h1>
            </div>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
              Table links
            </p>
          </div>

          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <div className="relative min-w-0 sm:w-80">
              <Search
                size={14}
                className="absolute left-3 top-1/2 -translate-y-1/2"
                style={{ color: 'var(--dash-text-dim)' }}
              />
              <input
                className="h-9 w-full rounded-lg pl-9 pr-3 text-sm outline-none transition-colors focus:border-[var(--dash-border-focus)]"
                style={{
                  background: 'var(--dash-input-bg)',
                  border: '1px solid var(--dash-border)',
                  color: 'var(--dash-text)',
                }}
                placeholder="Search tables"
                value={searchTerm}
                onChange={(event) => setSearchTerm(event.target.value)}
              />
            </div>
            <Button size="sm" onClick={() => navigate({ to: '/qr-codes/create' })}>
              <Plus size={14} />
              Custom QR Generator
            </Button>
          </div>
        </div>

        <div
          className="grid overflow-hidden rounded-lg border sm:grid-cols-3"
          style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-surface)' }}
        >
          {[
            ['Tables', String(tables.length)],
            ['Ready', String(readyTables)],
            ['Seats', String(totalCapacity)],
          ].map(([label, value], index) => (
            <div
              key={label}
              className="px-4 py-3"
              style={{
                borderLeft: index === 0 ? undefined : '1px solid var(--dash-border)',
                borderTop: index === 0 ? undefined : '1px solid var(--dash-border)',
              }}
            >
              <p
                className="text-[11px] font-medium uppercase"
                style={{ color: 'var(--dash-text-dim)' }}
              >
                {label}
              </p>
              <p className="mt-1 text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                {value}
              </p>
            </div>
          ))}
        </div>

        <section
          className="w-full overflow-hidden rounded-lg border"
          style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-surface)' }}
        >
          <div
            className="hidden px-4 py-2.5 text-[11px] font-semibold uppercase md:grid md:grid-cols-[minmax(0,1fr)_100px_140px_120px] md:items-center"
            style={{
              color: 'var(--dash-text-dim)',
              borderBottom: '1px solid var(--dash-border)',
              background: 'var(--dash-hover)',
            }}
          >
            <span>Table</span>
            <span>Seats</span>
            <span>Status</span>
            <span className="text-right">Actions</span>
          </div>

          {isLoading ? (
            <div>
              {[0, 1, 2, 3].map((row) => (
                <div
                  key={row}
                  className="grid gap-3 px-4 py-4 md:grid-cols-[minmax(0,1fr)_100px_140px_120px] md:items-center"
                  style={{ borderBottom: row === 3 ? undefined : '1px solid var(--dash-border)' }}
                >
                  <div
                    className="h-3 w-44 animate-pulse rounded"
                    style={{ background: 'var(--dash-hover)' }}
                  />
                  <div
                    className="h-3 w-10 animate-pulse rounded"
                    style={{ background: 'var(--dash-hover)' }}
                  />
                  <div
                    className="h-6 w-24 animate-pulse rounded-full"
                    style={{ background: 'var(--dash-hover)' }}
                  />
                  <div
                    className="ml-auto h-7 w-16 animate-pulse rounded-md"
                    style={{ background: 'var(--dash-hover)' }}
                  />
                </div>
              ))}
            </div>
          ) : filteredTables.length === 0 ? (
            <div className="flex min-h-64 flex-col items-center justify-center px-4 py-16 text-center">
              <div
                className="mb-4 flex h-12 w-12 items-center justify-center rounded-lg"
                style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
              >
                <QrCodeIcon size={22} />
              </div>
              <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                No tables found
              </p>
              <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                {searchTerm ? 'Try another search.' : 'No QR codes yet.'}
              </p>
              {!searchTerm && (
                <Button
                  className="mt-4"
                  size="sm"
                  variant="outline"
                  onClick={() => navigate({ to: '/qr-codes/create' })}
                >
                  <Plus size={14} />
                  QR Code
                </Button>
              )}
            </div>
          ) : (
            <div>
              {filteredTables.map((table, index) => {
                const hasQrCode = Boolean(table.qrCodeUrl);
                return (
                  <div
                    key={table.id}
                    className="grid gap-3 px-4 py-4 transition-colors hover:bg-[var(--dash-hover)] md:grid-cols-[minmax(0,1fr)_100px_140px_120px] md:items-center"
                    style={{
                      borderBottom:
                        index === filteredTables.length - 1
                          ? undefined
                          : '1px solid var(--dash-border)',
                    }}
                  >
                    <div className="min-w-0">
                      <p
                        className="truncate text-sm font-semibold"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        {table.name}
                      </p>
                      <p
                        className="mt-0.5 truncate text-xs"
                        style={{ color: 'var(--dash-text-dim)' }}
                      >
                        {hasQrCode ? 'QR ready' : 'Needs QR code'}
                      </p>
                    </div>
                    <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                      {table.capacity || 0}
                    </p>
                    <div>
                      <Badge variant={table.status === 'available' ? 'success' : 'outline'}>
                        {formatStatus(table.status || 'unknown')}
                      </Badge>
                    </div>
                    <div className="flex justify-start gap-2 md:justify-end">
                      <button
                        className="flex h-8 w-8 items-center justify-center rounded-md transition-colors hover:bg-[var(--dash-hover)] disabled:pointer-events-none disabled:opacity-35"
                        style={{
                          color: 'var(--dash-text-muted)',
                          border: '1px solid var(--dash-border)',
                        }}
                        aria-label="Copy QR URL"
                        title="Copy QR URL"
                        disabled={!hasQrCode}
                        onClick={() => {
                          void navigator.clipboard.writeText(table.qrCodeUrl);
                          toast.success('URL copied');
                        }}
                      >
                        <Copy size={14} />
                      </button>
                      <a
                        href={table.qrCodeUrl || undefined}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex h-8 w-8 items-center justify-center rounded-md transition-colors hover:bg-[var(--dash-hover)] aria-disabled:pointer-events-none aria-disabled:opacity-35"
                        style={{
                          color: 'var(--dash-text-muted)',
                          border: '1px solid var(--dash-border)',
                        }}
                        aria-label="Open QR URL"
                        aria-disabled={!hasQrCode}
                        title="Open QR URL"
                      >
                        <ExternalLink size={14} />
                      </a>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>
      </div>
    </>
  );
};

export default QrCodes;
