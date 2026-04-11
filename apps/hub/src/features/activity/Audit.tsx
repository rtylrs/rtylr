import React, { useMemo, useState } from 'react';
import { ScrollText, Download, Search } from 'lucide-react';
import { Button, Input } from "../../shared/components/ui";
import { ActivityEntryRow, Breadcrumbs, EmptyState, SkeletonList } from "../../shared/components/hub";
import { useAuditLog } from "../../shared/hooks/queries/useActivity";
import { toast } from 'sonner';

const Audit: React.FC = () => {
  const { data = [], isLoading } = useAuditLog(500);
  const [query, setQuery] = useState('');
  const [severity, setSeverity] = useState<'all' | 'info' | 'warning' | 'critical'>('all');

  const filtered = useMemo(() => {
    return data.filter((e) => {
      const matchesQuery =
        !query.trim() ||
        e.actor.toLowerCase().includes(query.toLowerCase()) ||
        e.action.toLowerCase().includes(query.toLowerCase()) ||
        e.target.toLowerCase().includes(query.toLowerCase()) ||
        e.ip?.toLowerCase().includes(query.toLowerCase());
      const matchesSeverity = severity === 'all' || e.severity === severity;
      return matchesQuery && matchesSeverity;
    });
  }, [data, query, severity]);

  const exportCsv = () => {
    const rows = [
      ['id', 'actor', 'action', 'target', 'app', 'ip', 'severity', 'created_at'],
      ...filtered.map((e) => [e.id, e.actor, e.action, e.target, e.app, e.ip ?? '', e.severity ?? '', e.createdAt]),
    ];
    const csv = rows.map((r) => r.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `rtylr-audit-${Date.now()}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success('Audit log exported');
  };

  return (
    <div className="space-y-5 max-w-5xl">
      
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="h-9 w-9 rounded-xl flex items-center justify-center" style={{ background: 'rgba(239,68,68,0.15)' }}>
            <ScrollText size={16} style={{ color: '#ef4444' }} />
          </div>
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Audit log</h1>
            <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
              Tamper-evident history of security-relevant events.
            </p>
          </div>
        </div>
        <Button variant="secondary" onClick={exportCsv}>
          <Download size={14} />
          Export CSV
        </Button>
      </div>

      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 max-w-sm">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" style={{ color: 'var(--hub-text-dim)' }} />
          <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search actor, action, IP..." className="pl-9" />
        </div>
        <div className="flex items-center gap-1">
          {(['all', 'info', 'warning', 'critical'] as const).map((s) => (
            <button
              key={s}
              onClick={() => setSeverity(s)}
              className="px-2.5 py-1 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-colors"
              style={{
                background: severity === s ? 'rgba(239,68,68,0.15)' : 'var(--hub-hover)',
                color:      severity === s ? '#ef4444' : 'var(--hub-text-muted)',
                border:     severity === s ? '1px solid rgba(239,68,68,0.3)' : '1px solid var(--hub-border)',
              }}
            >
              {s}
            </button>
          ))}
        </div>
      </div>

      <div className="rounded-xl overflow-hidden" style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}>
        {isLoading ? (
          <div className="p-2"><SkeletonList rows={10} /></div>
        ) : filtered.length === 0 ? (
          <div className="p-4">
            <EmptyState icon={ScrollText} title="No audit entries" description="Security-relevant events will appear here." />
          </div>
        ) : (
          <div className="p-2">
            {filtered.map((e) => <ActivityEntryRow key={e.id} entry={e} showIp />)}
          </div>
        )}
      </div>
    </div>
  );
};

export default Audit;
