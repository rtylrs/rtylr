import React from 'react';
import { Download, Upload, FileJson } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '../../shared/components/ui';
import { Breadcrumbs, SettingSection } from '../../shared/components/dash';

const DataExport: React.FC = () => {
  const exportAll = (): void => {
    toast.info('Full workspace export coming soon');
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
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            Export data
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
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
        <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
          We&apos;ll generate a downloadable file with all your workspace data. Large workspaces may
          take a few minutes.
        </p>
      </SettingSection>

      <SettingSection
        title="Scheduled exports"
        description="Automatic exports to your S3 bucket or email."
      >
        <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
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
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            Import data
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
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
            style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
          >
            <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
              {src.name}
            </p>
            <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
              {src.description}
            </p>
            <Button className="mt-3 w-full" variant="secondary" size="sm" disabled>
              Start import
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
};

export default DataExport;
