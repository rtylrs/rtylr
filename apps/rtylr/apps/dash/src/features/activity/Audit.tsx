import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { ScrollText, Download, Search, ChevronLeft, ChevronRight } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { Button, Input } from '../../shared/components/ui';
import { ActivityEntryRow, SkeletonList } from '../../shared/components/dash';
import { useAuditLog } from '../../shared/hooks/queries/useActivity';
import { toast } from 'sonner';

const PAGE_SIZE = 50;

const Audit: React.FC = () => {
  const { t } = useTranslation();
  const [page, setPage] = useState(1);
  const { data, isLoading, isFetching, refetch } = useAuditLog({ page, pageSize: PAGE_SIZE });
  const [query, setQuery] = useState('');
  const [severity, setSeverity] = useState<'all' | 'info' | 'warning' | 'critical'>('all');
  const isFiltering = query.trim() !== '' || severity !== 'all';

  const entries = useMemo(() => data?.entries ?? [], [data]);
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const rangeStart = total === 0 ? 0 : (page - 1) * PAGE_SIZE + 1;
  const rangeEnd = Math.min(page * PAGE_SIZE, total);

  // Search + severity refine the current page; pagination is server-side,
  // so changing either resets to the first page.
  const filtered = useMemo(() => {
    return entries.filter((e) => {
      const matchesQuery =
        !query.trim() ||
        e.actor.toLowerCase().includes(query.toLowerCase()) ||
        e.action.toLowerCase().includes(query.toLowerCase()) ||
        e.target.toLowerCase().includes(query.toLowerCase()) ||
        e.ip?.toLowerCase().includes(query.toLowerCase());
      const matchesSeverity = severity === 'all' || e.severity === severity;
      return matchesQuery && matchesSeverity;
    });
  }, [entries, query, severity]);

  const exportCsv = () => {
    const rows = [
      ['id', 'actor', 'action', 'target', 'app', 'ip', 'severity', 'created_at'],
      ...filtered.map((e) => [
        e.id,
        e.actor,
        e.action,
        e.target,
        e.app,
        e.ip ?? '',
        e.severity ?? '',
        e.createdAt,
      ]),
    ];
    const csv = rows
      .map((r) => r.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(','))
      .join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `rtylr-audit-page-${page}-${Date.now()}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success(`Exported page ${page} (${filtered.length} entries)`);
  };

  return (
    <>
      <HelmetContainer title="Audit log · rtylr" />
      <div className="space-y-5 max-w-5xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div
              className="h-9 w-9 rounded-xl flex items-center justify-center"
              style={{ background: 'rgba(239,68,68,0.15)' }}
            >
              <ScrollText size={16} style={{ color: '#ef4444' }} />
            </div>
            <div>
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                Audit log
              </h1>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                Tamper-evident history of security-relevant events.
              </p>
            </div>
          </div>
          <Button variant="secondary" onClick={exportCsv}>
            <Download size={14} />
            Export CSV (page {page})
          </Button>
        </div>

        <div className="flex items-center gap-3 flex-wrap">
          <div className="relative flex-1 max-w-sm">
            <Search
              size={14}
              className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"
              style={{ color: 'var(--dash-text-dim)' }}
            />
            <Input
              value={query}
              onChange={(e) => {
                setQuery(e.target.value);
                setPage(1);
              }}
              placeholder="Search actor, action, IP..."
              className="pl-9"
            />
          </div>
          <div className="flex items-center gap-1">
            {(['all', 'info', 'warning', 'critical'] as const).map((s) => (
              <button
                type="button"
                key={s}
                onClick={() => {
                  setSeverity(s);
                  setPage(1);
                }}
                className="px-2.5 py-1 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-colors"
                style={{
                  background: severity === s ? 'rgba(239,68,68,0.15)' : 'var(--dash-hover)',
                  color: severity === s ? '#ef4444' : 'var(--dash-text-muted)',
                  border:
                    severity === s
                      ? '1px solid rgba(239,68,68,0.3)'
                      : '1px solid var(--dash-border)',
                }}
              >
                {s}
              </button>
            ))}
          </div>
        </div>

        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          {isLoading ? (
            <div className="p-2">
              <SkeletonList rows={10} />
            </div>
          ) : filtered.length === 0 ? (
            <div className="p-4">
              <EmptyState
                icon={ScrollText}
                title={t('activity.audit.emptyTitle')}
                description={t('activity.audit.emptyDescription')}
                action={{ label: t('common.refresh'), onClick: () => void refetch() }}
              />
            </div>
          ) : (
            <div className="p-2">
              {filtered.map((e, i) => (
                <ActivityEntryRow key={e.id} entry={e} index={i} showIp />
              ))}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between">
          <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
            {total === 0
              ? 'No entries'
              : isFiltering
                ? t(
                    'activity.audit.filteredShowing',
                    'Showing {{shown}} of {{total}} on this page',
                    {
                      shown: filtered.length,
                      total,
                    }
                  )
                : `Showing ${rangeStart}–${rangeEnd} of ${total}`}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              disabled={page <= 1 || isFetching}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              <ChevronLeft size={14} />
              Previous
            </Button>
            <span className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              Page {page} of {totalPages}
            </span>
            <Button
              variant="secondary"
              disabled={page >= totalPages || isFetching}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            >
              Next
              <ChevronRight size={14} />
            </Button>
          </div>
        </div>
      </div>
    </>
  );
};

export default Audit;
