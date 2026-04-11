import React from 'react';
import { Download, Upload, FileJson } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '../../shared/components/ui';
import { Breadcrumbs, SettingSection } from '../../shared/components/hub';

const DataExport: React.FC = () => {
  const exportAll = () => {
    const snapshot = {
      exportedAt: new Date().toISOString(),
      message: 'Snapshot export placeholder',
    };
    const blob = new Blob([JSON.stringify(snapshot, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `rtylr-export-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success('Export started');
  };

  return (
    <div className="space-y-5 max-w-3xl">
      <div className="flex items-center gap-3">
        <div
          className="h-9 w-9 rounded-xl flex items-center justify-center"
          style={{ background: 'rgba(69,131,196,0.15)' }}
        >
          <FileJson size={16} style={{ color: '#4583C4' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Export data
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
            Download a full snapshot of your workspace data.
          </p>
        </div>
      </div>

      <SettingSection
        title="Full workspace export"
        description="Orders, customers, inventory, team, settings. JSON format."
        footer={
          <Button size="sm" onClick={exportAll}>
            <Download size={13} />
            Export everything
          </Button>
        }
      >
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
          We&apos;ll generate a downloadable file with all your workspace data. Large workspaces may
          take a few minutes.
        </p>
      </SettingSection>

      <SettingSection
        title="Scheduled exports"
        description="Automatic exports to your S3 bucket or email."
      >
        <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
          Scheduled exports are available on the Pro plan.
        </p>
      </SettingSection>
    </div>
  );
};

export const DataImport: React.FC = () => {
  return (
    <div className="space-y-5 max-w-3xl">
      <div className="flex items-center gap-3">
        <div
          className="h-9 w-9 rounded-xl flex items-center justify-center"
          style={{ background: 'rgba(16,185,129,0.15)' }}
        >
          <Upload size={16} style={{ color: '#10b981' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Import data
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
            Bring your existing data to rtylr.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {[
          { name: 'Square', description: 'Import catalog, customers, and sales history.' },
          { name: 'Shopify', description: 'Pull products, orders, and customers.' },
          { name: 'Toast', description: 'Migrate menus and team from Toast.' },
          { name: 'Clover', description: 'Import inventory and sales from Clover.' },
        ].map((src) => (
          <div
            key={src.name}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {src.name}
            </p>
            <p className="text-xs mt-1" style={{ color: 'var(--hub-text-muted)' }}>
              {src.description}
            </p>
            <Button
              className="mt-3 w-full"
              variant="secondary"
              size="sm"
              onClick={() => toast.info(`${src.name} import coming soon`)}
            >
              Start import
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
};

export default DataExport;
