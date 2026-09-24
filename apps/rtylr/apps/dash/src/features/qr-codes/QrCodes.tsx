import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { Plus, QrCode as QrCodeIcon, RefreshCw, Search } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { EmptyState } from '@voxire/ui';
import { Badge, Button } from '../../shared/components/ui';
import { configClient, tableClient } from '../../shared/lib/clients';
import { useUserStore } from '../../shared/stores/userStore';

interface QrTable {
  id: number;
  name: string;
  capacity: number;
  status: string;
}

const useQrTables = () =>
  useQuery<QrTable[]>({
    queryKey: ['qr-tables'],
    queryFn: async () => {
      const res = await tableClient.listTables({});
      return (res.tables ?? []).map((table) => ({
        id: Number(table.id ?? 0),
        name: table.label ?? '',
        capacity: table.seats ?? 0,
        status: table.status ?? '',
      }));
    },
  });

const useBusinessSlug = (businessId: number) =>
  useQuery<string>({
    queryKey: ['qr-business-slug', businessId],
    enabled: businessId > 0,
    queryFn: async () => {
      const response = await configClient.getConfig({});
      return response.business?.slug?.trim() ?? '';
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
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState('');
  const { data: tables = [], isLoading, isError, refetch } = useQrTables();
  const user = useUserStore((state) => state.user);
  const setUser = useUserStore((state) => state.setUser);
  const businessId = Number(user?.businessId ?? 0);
  const { data: configSlugData } = useBusinessSlug(businessId);
  const configSlug = typeof configSlugData === 'string' ? configSlugData.trim() : '';
  const businessSlug = user?.businessSlug?.trim() || user?.business_slug?.trim() || configSlug;
  const canOpenGenerator = !user || Boolean(businessSlug);
  const normalizedSearch = searchTerm.trim().toLowerCase();

  const filteredTables = useMemo(
    () =>
      tables.filter((table) =>
        normalizedSearch ? table.name.toLowerCase().includes(normalizedSearch) : true
      ),
    [normalizedSearch, tables]
  );

  const totalCapacity = tables.reduce((sum, table) => sum + (table.capacity || 0), 0);

  const handleOpenGenerator = (): void => {
    if (!canOpenGenerator) return;
    if (businessSlug && setUser) {
      if (!user) {
        setUser({ businessSlug });
      } else if (user.businessSlug !== businessSlug && user.business_slug !== businessSlug) {
        setUser({ ...user, businessSlug });
      }
    }
    navigate({ to: '/qr-codes/create' });
  };

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
            <Button
              size="sm"
              onClick={handleOpenGenerator}
              disabled={!canOpenGenerator}
              title={
                !canOpenGenerator
                  ? 'Set a business slug before generating a menu QR code.'
                  : undefined
              }
            >
              <Plus size={14} />
              Custom QR Generator
            </Button>
          </div>
        </div>

        {!isLoading && !isError && (
          <div
            className="grid overflow-hidden rounded-lg border sm:grid-cols-2"
            style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-surface)' }}
          >
            {[
              ['Tables', String(tables.length)],
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
        )}

        <section
          className="w-full overflow-hidden rounded-lg border"
          style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-surface)' }}
        >
          <div
            className="hidden px-4 py-2.5 text-[11px] font-semibold uppercase md:grid md:grid-cols-[minmax(0,1fr)_100px_140px] md:items-center"
            style={{
              color: 'var(--dash-text-dim)',
              borderBottom: '1px solid var(--dash-border)',
              background: 'var(--dash-hover)',
            }}
          >
            <span>Table</span>
            <span>Seats</span>
            <span>Status</span>
          </div>

          {isLoading ? (
            <div>
              {[0, 1, 2, 3].map((row) => (
                <div
                  key={row}
                  className="grid gap-3 px-4 py-4 md:grid-cols-[minmax(0,1fr)_100px_140px] md:items-center"
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
                </div>
              ))}
            </div>
          ) : isError ? (
            <EmptyState
              icon={RefreshCw}
              title="Unable to load QR codes"
              description="The table list could not be loaded."
              action={{ label: t('common.refresh'), onClick: () => void refetch() }}
            />
          ) : filteredTables.length === 0 ? (
            <EmptyState
              icon={QrCodeIcon}
              title={t('qrCodes.emptyNoTablesTitle')}
              description={
                searchTerm ? t('qrCodes.emptyTryAnotherSearch') : t('qrCodes.emptyNoQrCodesYet')
              }
              action={
                searchTerm
                  ? { label: t('common.clearSearch'), onClick: () => setSearchTerm('') }
                  : canOpenGenerator
                    ? {
                        label: t('qrCodes.createQrCode'),
                        onClick: handleOpenGenerator,
                      }
                    : { label: t('common.refresh'), onClick: () => void refetch() }
              }
            />
          ) : (
            <div>
              {filteredTables.map((table, index) => {
                return (
                  <div
                    key={table.id}
                    className="grid gap-3 px-4 py-4 transition-colors hover:bg-[var(--dash-hover)] md:grid-cols-[minmax(0,1fr)_100px_140px] md:items-center"
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
                    </div>
                    <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                      {table.capacity || 0}
                    </p>
                    <div>
                      <Badge variant={table.status === 'available' ? 'success' : 'outline'}>
                        {formatStatus(table.status || 'unknown')}
                      </Badge>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>

        <p className="px-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
          {t('qrCodes.tableQrComingSoon')}
        </p>
      </div>
    </>
  );
};

export default QrCodes;
