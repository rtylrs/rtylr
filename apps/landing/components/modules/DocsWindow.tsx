import React from 'react';
import {
  Rocket,
  Utensils,
  Package,
  Users,
  UserCheck,
  DollarSign,
  GitMerge,
  LucideProps,
} from 'lucide-react';

const DOCS_SECTIONS: { icon: React.FC<LucideProps>; title: string; desc: string }[] = [
  { icon: Rocket, title: 'Getting Started', desc: 'Set up your workspace in minutes' },
  { icon: Utensils, title: 'POS', desc: 'Orders, tables, kitchen display' },
  { icon: Package, title: 'ERP', desc: 'Inventory, suppliers, purchase orders' },
  { icon: Users, title: 'CRM', desc: 'Contacts, pipelines, campaigns' },
  { icon: UserCheck, title: 'HR', desc: 'Staff, shifts, payroll, leave' },
  { icon: DollarSign, title: 'Finance', desc: 'Invoices, expenses, reports' },
  { icon: GitMerge, title: 'Flow', desc: 'Projects, tasks, automations' },
];

export default function DocsWindow(): React.JSX.Element {
  return (
    <div
      style={{
        height: '100%',
        background: '#151515',
        overflowY: 'auto',
        padding: '20px 24px',
      }}
    >
      <div
        style={{
          fontSize: 11,
          fontWeight: 600,
          letterSpacing: '0.1em',
          color: 'rgba(255,255,255,0.3)',
          textTransform: 'uppercase',
          marginBottom: 20,
        }}
      >
        Documentation
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {DOCS_SECTIONS.map((s) => {
          const Icon = s.icon;
          return (
            <div
              key={s.title}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 14,
                padding: '12px 16px',
                background: 'rgba(255,255,255,0.03)',
                border: '1px solid rgba(255,255,255,0.06)',
                borderRadius: 8,
                cursor: 'pointer',
                transition: 'background 0.15s',
              }}
              onMouseEnter={(e) =>
                ((e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.06)')
              }
              onMouseLeave={(e) =>
                ((e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.03)')
              }
            >
              <Icon size={16} color="rgba(255,255,255,0.45)" strokeWidth={1.5} />
              <div>
                <div style={{ fontSize: 13, fontWeight: 500, color: '#e8e6e1' }}>{s.title}</div>
                <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)', marginTop: 2 }}>
                  {s.desc}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
