import React, { useRef, useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { getOneKitErrorMessage, uploadAsset } from '@voxire/api';
import { APP_URLS } from '@voxire/config';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { RotateCcw, Save, Image as ImageIcon, Settings2, Users, AlertTriangle } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import { useSettings } from '../../shared/lib/settings';
import { CURRENCY_OPTIONS, type Currency } from '../../shared/lib/money';
import {
  cardStyle,
  inputStyle,
  labelStyle,
  primaryBtn,
  ghostBtn,
  ACCENT,
} from '../../shared/lib/tokens';

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

type TabId = 'general' | 'team';

interface Tab {
  id: TabId;
  labelKey: string;
  icon: React.ComponentType<{ size?: number }>;
}

const TABS: Tab[] = [
  { id: 'general', labelKey: 'settings.tabs.general', icon: Settings2 },
  { id: 'team', labelKey: 'settings.tabs.team', icon: Users },
];

// ── General Tab ───────────────────────────────────────────────────────────────

const GeneralTab: React.FC = () => {
  const { t } = useTranslation();
  const settings = useSettings();
  const [uploadingLogo, setUploadingLogo] = useState(false);
  const logoInputRef = useRef<HTMLInputElement>(null);

  const handleLogoUpload = async (file?: File | null): Promise<void> => {
    if (!file) return;
    setUploadingLogo(true);
    try {
      const uploaded = await uploadAsset(file);
      settings.setSetting('logoUrl', uploaded.url);
    } catch (err) {
      toast.error(getOneKitErrorMessage(err, 'Upload failed'));
    } finally {
      setUploadingLogo(false);
      if (logoInputRef.current) logoInputRef.current.value = '';
    }
  };

  const handleReset = (): void => {
    settings.reset();
    toast.success(t('settings.reset'));
  };

  const handleSave = (): void => {
    try {
      // Settings are persisted locally via Zustand persist middleware on every setSetting call.
      // This button confirms the current state is intentionally saved.
      toast.success(t('settings.saved'));
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, 'Failed to save settings'));
    }
  };

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginBottom: 20 }}>
        <button type="button" style={ghostBtn} onClick={handleReset}>
          <RotateCcw size={13} /> Reset
        </button>
        <button type="button" style={primaryBtn} onClick={handleSave}>
          <Save size={13} /> Save
        </button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        <Card title={t('settings.company.heading')}>
          <Field label={t('settings.company.name')}>
            <input
              value={settings.companyName}
              onChange={(e) => settings.setSetting('companyName', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label={t('settings.company.billingEmail')}>
            <input
              value={settings.companyEmail}
              onChange={(e) => settings.setSetting('companyEmail', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label={t('settings.company.brandColor')}>
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
          <Field label={t('settings.company.logoUrl')}>
            <div style={{ display: 'flex', gap: 8 }}>
              <input
                value={settings.logoUrl}
                onChange={(e) => settings.setSetting('logoUrl', e.target.value)}
                placeholder="https://..."
                style={inputStyle}
              />
              <button
                type="button"
                style={{ ...ghostBtn, padding: '8px 12px' }}
                disabled={uploadingLogo}
                onClick={() => logoInputRef.current?.click()}
                aria-label={t('settings.company.uploadLogo')}
              >
                <ImageIcon size={13} />
              </button>
              <input
                ref={logoInputRef}
                type="file"
                accept="image/*"
                style={{ display: 'none' }}
                onChange={(e) => {
                  void handleLogoUpload(e.target.files?.[0]);
                }}
              />
            </div>
          </Field>
        </Card>

        <Card title={t('settings.invoicing.heading')}>
          <Field label={t('settings.invoicing.prefix')}>
            <input
              value={settings.invoicePrefix}
              onChange={(e) => settings.setSetting('invoicePrefix', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label={t('settings.invoicing.defaultTaxRate')}>
            <input
              type="number"
              step="0.1"
              value={settings.defaultTaxRate}
              onChange={(e) => {
                const v = parseFloat(e.target.value);
                if (!isNaN(v) && v >= 0) settings.setSetting('defaultTaxRate', v);
              }}
              style={inputStyle}
            />
          </Field>
          <Field label={t('settings.invoicing.lateFee')}>
            <input
              type="number"
              step="0.1"
              value={settings.lateFeePercent}
              onChange={(e) => {
                const v = parseFloat(e.target.value);
                if (!isNaN(v) && v >= 0) settings.setSetting('lateFeePercent', v);
              }}
              style={inputStyle}
            />
          </Field>
          <Field label={t('settings.invoicing.gracePeriod')}>
            <input
              type="number"
              value={settings.lateFeeGraceDays}
              onChange={(e) => {
                const v = parseInt(e.target.value, 10);
                if (!isNaN(v) && v >= 0) settings.setSetting('lateFeeGraceDays', v);
              }}
              style={inputStyle}
            />
          </Field>
          <Field label={t('settings.invoicing.defaultNotes')}>
            <textarea
              value={settings.invoiceNotes}
              onChange={(e) => settings.setSetting('invoiceNotes', e.target.value)}
              rows={3}
              style={{ ...inputStyle, resize: 'vertical', fontFamily: 'inherit' }}
            />
          </Field>
        </Card>

        <Card title={t('settings.accounting.heading')}>
          <Field label={t('settings.accounting.currency')}>
            <select
              value={settings.currency}
              onChange={(e) => settings.setSetting('currency', e.target.value as Currency)}
              style={inputStyle}
            >
              {CURRENCY_OPTIONS.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </Field>
          <Field label={t('settings.accounting.fiscalYearStart')}>
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

        <Card title={t('settings.shortcuts.heading')}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <ShortcutRow label={t('settings.shortcuts.globalSearch')} keys={['cmd', 'k']} />
            <ShortcutRow label={t('settings.shortcuts.newInvoice')} keys={['n']} />
            <ShortcutRow label={t('settings.shortcuts.newExpense')} keys={['e']} />
            <ShortcutRow label={t('settings.shortcuts.focusSearch')} keys={['/']} />
          </div>
        </Card>
      </div>
    </>
  );
};

// ── Team Tab ──────────────────────────────────────────────────────────────────

const TeamTab: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 600 }}>
      <div>
        <div style={{ fontSize: 15, fontWeight: 600, color: 'var(--hub-text)', marginBottom: 4 }}>
          {t('settings.team.accessTitle')}
        </div>
        <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
          {t('settings.team.accessDescription')}
        </div>
      </div>
      <a
        href={`${APP_URLS.dash}/team`}
        style={{
          display: 'inline-flex',
          alignSelf: 'flex-start',
          padding: '8px 12px',
          borderRadius: 8,
          background: ACCENT,
          color: '#fff',
          fontSize: 12,
          fontWeight: 600,
          textDecoration: 'none',
        }}
      >
        {t('settings.team.accessAction')}
      </a>
    </div>
  );
};

// ── Main Settings Component ───────────────────────────────────────────────────

const Settings: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<TabId>('general');

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('nav.settings'))} />
      <PageHeader title={t('settings.title')} />

      {/* Tab bar */}
      <div
        style={{
          display: 'flex',
          gap: 0,
          borderBottom: '1px solid var(--hub-border)',
        }}
      >
        {TABS.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '10px 16px',
                fontSize: 13,
                fontWeight: isActive ? 600 : 500,
                color: isActive ? ACCENT : 'var(--hub-text-muted)',
                background: 'transparent',
                border: 'none',
                borderBottom: isActive ? `2px solid ${ACCENT}` : '2px solid transparent',
                marginBottom: -1,
                cursor: 'pointer',
                transition: 'color 0.15s, border-color 0.15s',
              }}
            >
              <Icon size={14} />
              {t(tab.labelKey)}
            </button>
          );
        })}
      </div>

      {/* Tab content */}
      <motion.div
        key={activeTab}
        initial={{ opacity: 0, y: 6 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.15 }}
      >
        {activeTab === 'general' && <GeneralTab />}
        {activeTab === 'team' && <TeamTab />}
      </motion.div>
    </motion.div>
  );
};

// ── Shared sub-components ─────────────────────────────────────────────────────

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
