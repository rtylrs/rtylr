import React from 'react';
import { HelmetContainer } from '@voxire/components';

interface ChangelogEntry {
  version: string;
  date: string;
  changes: string[];
}

const entries: ChangelogEntry[] = [
  {
    version: '1.6.0',
    date: '2026-06-06',
    changes: [
      'Caret — your AI business assistant — is now available across Dash, ERP, HR, CRM, Finance, Flow, and Insights',
      'Ask questions about your business in plain language; Analyst mode answers from your live data',
      'Streaming responses with conversation history and memory',
      'Monthly AI usage allowance per plan, with shared team pool or per-person allocation (Settings → AI Usage)',
    ],
  },
  {
    version: '1.5.0',
    date: '2026-05-12',
    changes: [
      'ERP: richer products — pricing, tax, weight & dimensions, country of origin, lead time, and custom attributes',
      'ERP: product variant matrix — generate per-SKU price and stock from size/color options',
      'Flow: My Tasks is now a single board of everything assigned to you across all workspaces',
      'Dash: paginated, searchable audit log',
      'Faster, more reliable APIs backed by a dedicated read replica for analytics',
    ],
  },
  {
    version: '1.4.0',
    date: '2026-04-01',
    changes: [
      'Added Book app for appointment and booking management',
      'New webhook events for order status changes',
      'Improved API rate limiting with burst support',
      'Fixed currency formatting in Finance reports',
    ],
  },
  {
    version: '1.3.0',
    date: '2026-03-15',
    changes: [
      'Store app now supports custom domains',
      'Added bulk product import via CSV',
      'CRM contact merge functionality',
      'Performance improvements to POS checkout flow',
    ],
  },
  {
    version: '1.2.0',
    date: '2026-02-20',
    changes: [
      'Flow app with workflow automation and task boards',
      'Multi-location inventory tracking in ERP',
      'HR leave management and approval workflows',
      'New API endpoints for customer segments',
    ],
  },
  {
    version: '1.1.0',
    date: '2026-01-10',
    changes: [
      'Finance app with profit/loss and balance sheet reports',
      'Team roles and granular app permissions',
      'Addon marketplace for third-party extensions',
      'Improved search across all apps',
    ],
  },
  {
    version: '1.0.0',
    date: '2025-12-01',
    changes: [
      'Initial release of rtylr',
      'Hub, POS, ERP, HR, and CRM apps',
      'REST API with authentication and rate limiting',
      'Webhook support for real-time integrations',
    ],
  },
];

const Changelog: React.FC = () => {
  return (
    <>
      <HelmetContainer title="Changelog · rtylr" />
      <div>
        <h1 className="text-2xl font-bold mb-2" style={{ color: 'var(--docs-text)' }}>
          Changelog
        </h1>
        <p className="text-sm mb-8" style={{ color: 'var(--docs-text-muted)' }}>
          New features, improvements, and fixes for the rtylr platform.
        </p>

        <div className="flex flex-col gap-8">
          {entries.map((entry) => (
            <div
              key={entry.version}
              className="relative pl-6 border-l-2"
              style={{ borderColor: 'var(--docs-border)' }}
            >
              {/* Version dot */}
              <div
                className="absolute -left-[7px] top-0.5 w-3 h-3 rounded-full border-2"
                style={{ borderColor: 'var(--docs-accent)', background: 'var(--docs-bg)' }}
              />
              <div className="flex items-baseline gap-3 mb-2">
                <span className="text-base font-semibold" style={{ color: 'var(--docs-text)' }}>
                  v{entry.version}
                </span>
                <span className="text-xs" style={{ color: 'var(--docs-text-dim)' }}>
                  {entry.date}
                </span>
              </div>
              <ul className="list-disc pl-5 flex flex-col gap-1">
                {entry.changes.map((change, idx) => (
                  <li key={idx} className="text-sm" style={{ color: 'var(--docs-text)' }}>
                    {change}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>
    </>
  );
};

export default Changelog;
