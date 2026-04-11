import React, { useState } from 'react';
import { Flag } from 'lucide-react';
import { Breadcrumbs, PermissionCheckbox, SettingSection } from "../../shared/components/hub";
import { toast } from 'sonner';

interface Flag {
  key: string;
  label: string;
  description: string;
  enabled: boolean;
  stage: 'stable' | 'beta' | 'alpha';
}

const DEFAULT_FLAGS: Flag[] = [
  { key: 'hub.command_palette', label: 'Command palette',    description: 'Enable cmd+k command palette across Hub.',        enabled: true,  stage: 'stable' },
  { key: 'hub.onboarding',      label: 'Onboarding checklist', description: 'Show the onboarding checklist on the dashboard.', enabled: true, stage: 'stable' },
  { key: 'hub.welcome_banner',  label: 'Welcome banner',      description: 'First-run welcome banner on dashboard.',          enabled: true,  stage: 'stable' },
  { key: 'hub.kpi_live',        label: 'Live KPIs',           description: 'Poll KPIs every 30 seconds on the dashboard.',    enabled: false, stage: 'beta' },
  { key: 'hub.ai_assistant',    label: 'AI assistant',        description: 'Embed an AI helper in the topbar.',               enabled: false, stage: 'alpha' },
  { key: 'hub.theme_customizer',label: 'Theme customizer',    description: 'Customize brand colors and accent.',              enabled: true,  stage: 'beta' },
  { key: 'hub.marketplace',     label: 'App marketplace',     description: 'Show the third-party app marketplace.',           enabled: false, stage: 'alpha' },
];

const STORAGE_KEY = 'hub.flags';

const readFlags = (): Record<string, boolean> => {
  if (typeof window === 'undefined') return {};
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
  } catch {
    return {};
  }
};

const STAGE_META = {
  stable: { color: '#10b981', bg: 'rgba(16,185,129,0.15)' },
  beta:   { color: '#8b5cf6', bg: 'rgba(139,92,246,0.15)' },
  alpha:  { color: '#f59e0b', bg: 'rgba(245,158,11,0.15)' },
} as const;

const FeatureFlagsScreen: React.FC = () => {
  const [overrides, setOverrides] = useState<Record<string, boolean>>(readFlags);

  const flags = DEFAULT_FLAGS.map((f) => ({
    ...f,
    enabled: overrides[f.key] ?? f.enabled,
  }));

  const toggle = (key: string, value: boolean) => {
    const next = { ...overrides, [key]: value };
    setOverrides(next);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    toast.success(`Flag ${key} ${value ? 'enabled' : 'disabled'}`);
  };

  return (
    <div className="space-y-5 max-w-3xl">
      
      <div className="flex items-center gap-3">
        <div className="h-9 w-9 rounded-xl flex items-center justify-center" style={{ background: 'rgba(245,158,11,0.15)' }}>
          <Flag size={16} style={{ color: '#f59e0b' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Feature flags</h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
            Toggle experimental features. Overrides are local to your browser.
          </p>
        </div>
      </div>

      <SettingSection title="Flags" description={`${flags.filter((f) => f.enabled).length} of ${flags.length} enabled`}>
        <div className="space-y-1 -mx-2">
          {flags.map((f) => {
            const meta = STAGE_META[f.stage];
            return (
              <div key={f.key} className="flex items-start gap-3 px-3 py-3 rounded-lg transition-colors hover:bg-[var(--hub-hover)]">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>{f.label}</p>
                    <span
                      className="px-1.5 py-0.5 rounded text-[9px] font-semibold uppercase"
                      style={{ background: meta.bg, color: meta.color }}
                    >
                      {f.stage}
                    </span>
                  </div>
                  <p className="text-[11px] mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{f.description}</p>
                  <code className="text-[10px] font-mono" style={{ color: 'var(--hub-text-dim)' }}>{f.key}</code>
                </div>
                <PermissionCheckbox checked={f.enabled} onChange={(v) => toggle(f.key, v)} />
              </div>
            );
          })}
        </div>
      </SettingSection>
    </div>
  );
};

export default FeatureFlagsScreen;
