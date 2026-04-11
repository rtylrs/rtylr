import React from 'react';

const entries = [
  { version: '0.2.0', date: 'Apr 10', text: 'Addon system, team invites, request access flow' },
  { version: '0.1.5', date: 'Apr 9', text: '75 backend modules migrated to sqlc' },
  {
    version: '0.1.0',
    date: 'Apr 9',
    text: 'Full platform launch: POS, ERP, HR, CRM, Finance, Flow',
  },
];

const ChangelogWidget: React.FC = () => {
  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-3" style={{ borderBottom: '1px solid var(--hub-border)' }}>
        <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
          What's new
        </p>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {entries.map((e) => (
          <div key={e.version} className="flex items-start gap-3">
            <span
              className="text-[10px] font-mono font-semibold px-1.5 py-0.5 rounded shrink-0 mt-0.5"
              style={{ background: 'rgba(69,131,196,0.1)', color: '#4583C4' }}
            >
              {e.version}
            </span>
            <div className="flex-1 min-w-0">
              <p className="text-xs leading-relaxed" style={{ color: 'var(--hub-text)' }}>
                {e.text}
              </p>
              <p className="text-[10px] mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
                {e.date}
              </p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ChangelogWidget;
