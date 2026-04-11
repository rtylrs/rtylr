import React from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { RotateCcw, Save, Image as ImageIcon } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import { useSettings } from '../../shared/lib/settings';
import { CURRENCY_OPTIONS } from '../../shared/lib/money';
import {
  cardStyle,
  inputStyle,
  labelStyle,
  primaryBtn,
  ghostBtn,
  ACCENT,
} from '../../shared/lib/tokens';

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

const Settings: React.FC = () => {
  const settings = useSettings();

  const handleReset = () => {
    settings.reset();
    toast.success('Settings reset');
  };

  const handleSave = () => {
    toast.success('Settings saved');
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <PageHeader
        title="Settings"
        actions={
          <>
            <button style={ghostBtn} onClick={handleReset}>
              <RotateCcw size={13} /> Reset
            </button>
            <button style={primaryBtn} onClick={handleSave}>
              <Save size={13} /> Save
            </button>
          </>
        }
      />

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        <Card title="Company">
          <Field label="Company name">
            <input
              value={settings.companyName}
              onChange={(e) => settings.setSetting('companyName', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label="Billing email">
            <input
              value={settings.companyEmail}
              onChange={(e) => settings.setSetting('companyEmail', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label="Brand color">
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <input
                type="color"
                value={settings.brandColor}
                onChange={(e) => settings.setSetting('brandColor', e.target.value)}
                style={{
                  width: 44,
                  height: 32,
                  border: 'none',
                  borderRadius: 8,
                  background: 'transparent',
                }}
              />
              <input
                value={settings.brandColor}
                onChange={(e) => settings.setSetting('brandColor', e.target.value)}
                style={inputStyle}
              />
            </div>
          </Field>
          <Field label="Logo URL">
            <div style={{ display: 'flex', gap: 8 }}>
              <input
                value={settings.logoUrl}
                onChange={(e) => settings.setSetting('logoUrl', e.target.value)}
                placeholder="https://..."
                style={inputStyle}
              />
              <button style={{ ...ghostBtn, padding: '8px 12px' }}>
                <ImageIcon size={13} />
              </button>
            </div>
          </Field>
        </Card>

        <Card title="Invoicing">
          <Field label="Invoice prefix">
            <input
              value={settings.invoicePrefix}
              onChange={(e) => settings.setSetting('invoicePrefix', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label="Default tax rate (%)">
            <input
              type="number"
              step="0.1"
              value={settings.defaultTaxRate}
              onChange={(e) =>
                settings.setSetting('defaultTaxRate', parseFloat(e.target.value) || 0)
              }
              style={inputStyle}
            />
          </Field>
          <Field label="Late fee (%)">
            <input
              type="number"
              step="0.1"
              value={settings.lateFeePercent}
              onChange={(e) =>
                settings.setSetting('lateFeePercent', parseFloat(e.target.value) || 0)
              }
              style={inputStyle}
            />
          </Field>
          <Field label="Grace period (days)">
            <input
              type="number"
              value={settings.lateFeeGraceDays}
              onChange={(e) =>
                settings.setSetting('lateFeeGraceDays', parseInt(e.target.value) || 0)
              }
              style={inputStyle}
            />
          </Field>
          <Field label="Default notes">
            <textarea
              value={settings.invoiceNotes}
              onChange={(e) => settings.setSetting('invoiceNotes', e.target.value)}
              rows={3}
              style={{ ...inputStyle, resize: 'vertical', fontFamily: 'inherit' }}
            />
          </Field>
        </Card>

        <Card title="Accounting">
          <Field label="Currency">
            <select
              value={settings.currency}
              onChange={(e) => settings.setSetting('currency', e.target.value as any)}
              style={inputStyle}
            >
              {CURRENCY_OPTIONS.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Fiscal year start">
            <select
              value={settings.fiscalYearStartMonth}
              onChange={(e) =>
                settings.setSetting('fiscalYearStartMonth', parseInt(e.target.value) || 1)
              }
              style={inputStyle}
            >
              {MONTHS.map((m, i) => (
                <option key={m} value={i + 1}>
                  {m}
                </option>
              ))}
            </select>
          </Field>
        </Card>

        <Card title="Shortcuts">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <ShortcutRow label="Global search" keys={['cmd', 'k']} />
            <ShortcutRow label="New invoice" keys={['n']} />
            <ShortcutRow label="New expense" keys={['e']} />
            <ShortcutRow label="Focus search" keys={['/']} />
          </div>
        </Card>
      </div>
    </motion.div>
  );
};

const Card: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <div style={{ ...cardStyle, padding: 20 }}>
    <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--hub-text)', marginBottom: 14 }}>
      {title}
    </div>
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>{children}</div>
  </div>
);

const Field: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
  <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
    <label style={labelStyle}>{label}</label>
    {children}
  </div>
);

const ShortcutRow: React.FC<{ label: string; keys: string[] }> = ({ label, keys }) => (
  <div
    style={{
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      padding: '6px 0',
      borderBottom: '1px solid var(--hub-border)',
    }}
  >
    <span style={{ fontSize: 13, color: 'var(--hub-text-muted)' }}>{label}</span>
    <div style={{ display: 'flex', gap: 4 }}>
      {keys.map((k) => (
        <kbd
          key={k}
          style={{
            padding: '3px 8px',
            borderRadius: 6,
            background: 'var(--hub-hover)',
            border: '1px solid var(--hub-border)',
            fontSize: 11,
            color: 'var(--hub-text)',
            textTransform: 'uppercase',
          }}
        >
          {k}
        </kbd>
      ))}
    </div>
  </div>
);

export default Settings;
