import React, { useRef, useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { getOneKitErrorMessage, uploadAsset } from '@voxire/api';
import { APP_URLS } from '@voxire/config';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { RotateCcw, Image as ImageIcon, Settings2, Users, AlertTriangle } from 'lucide-react';
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

const GeneralTab: React.FC = () => {
  const { t } = useTranslation('financeCommon');
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
    toast.success(t('financeSettings.reset'));
  };

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginBottom: 20 }}>
        <button type="button" style={ghostBtn} onClick={handleReset}>
          <RotateCcw size={13} /> {t('financeSettings.resetAction')}
        </button>
        <p className="text-sm text-[var(--dash-text-muted)]">
          {t('financeSettings.localPersistence')}
        </p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        <Card title={t('financeSettings.company.heading')}>
          <Field label={t('financeSettings.company.name')}>
            <input
              value={settings.companyName}
              onChange={(e) => settings.setSetting('companyName', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label={t('financeSettings.company.billingEmail')}>
            <input
              value={settings.companyEmail}
              onChange={(e) => settings.setSetting('companyEmail', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label={t('financeSettings.company.brandColor')}>
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
          <Field label={t('financeSettings.company.logoUrl')}>
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
                aria-label={t('financeSettings.company.uploadLogo')}
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

        <Card title={t('financeSettings.invoicing.heading')}>
          <Field label={t('financeSettings.invoicing.prefix')}>
            <input
              value={settings.invoicePrefix}
              onChange={(e) => settings.setSetting('invoicePrefix', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label={t('financeSettings.invoicing.defaultTaxRate')}>
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
          <Field label={t('financeSettings.invoicing.lateFee')}>
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
          <Field label={t('financeSettings.invoicing.gracePeriod')}>
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
          <Field label={t('financeSettings.invoicing.defaultNotes')}>
            <textarea
              value={settings.invoiceNotes}
              onChange={(e) => settings.setSetting('invoiceNotes', e.target.value)}
              rows={3}
              style={{ ...inputStyle, resize: 'vertical', fontFamily: 'inherit' }}
            />
          </Field>
        </Card>

        <Card title={t('financeSettings.accounting.heading')}>
          <Field label={t('financeSettings.accounting.currency')}>
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
          <Field label={t('financeSettings.accounting.fiscalYearStart')}>
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

        <Card title={t('financeSettings.shortcuts.heading')}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <ShortcutRow label={t('financeSettings.shortcuts.globalSearch')} keys={['cmd', 'k']} />
            <ShortcutRow label={t('financeSettings.shortcuts.newInvoice')} keys={['n']} />
            <ShortcutRow label={t('financeSettings.shortcuts.newExpense')} keys={['e']} />
            <ShortcutRow label={t('financeSettings.shortcuts.focusSearch')} keys={['/']} />
          </div>
        </Card>
      </div>
    </>
  );
};

export default GeneralTab;
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
