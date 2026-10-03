import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { TIMEZONES } from '@voxire/config';
import {
  companyProfileSchema,
  type CompanyProfileFormData,
} from '../../../shared/schemas/wizardSchemas';
import StepContainer from '../components/StepContainer';

interface CompanyProfileStepProps {
  onNext: (data: CompanyProfileFormData) => void;
  onBack?: () => void;
  isFirst?: boolean;
  isSubmitting?: boolean;
  sharedName?: string;
  initialValues?: Partial<CompanyProfileFormData>;
}

// The code is what gets stored; only the name is translated.
const SETUP_CURRENCIES = ['USD', 'EUR', 'GBP', 'AED', 'SAR', 'LBP', 'JPY', 'AUD', 'CAD'] as const;

const CompanyProfileStep = ({
  onNext,
  onBack,
  isFirst,
  isSubmitting,
  sharedName,
  initialValues,
}: CompanyProfileStepProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const [form, setForm] = useState<CompanyProfileFormData>({
    name: initialValues?.name ?? '',
    timezone: initialValues?.timezone ?? '',
    currency: initialValues?.currency ?? 'USD',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (): void => {
    const result = companyProfileSchema.safeParse({ ...form, name: sharedName || form.name });
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const field = issue.path[0];
        if (typeof field === 'string') {
          fieldErrors[field] = t('hrCommon.required');
        }
      }
      setErrors(fieldErrors);
      return;
    }
    setErrors({});
    onNext(result.data);
  };

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '10px 12px',
    borderRadius: '8px',
    border: '1px solid var(--hr-border)',
    background: 'var(--hr-input-bg)',
    color: 'var(--hr-text)',
    fontSize: '16px',
    outline: 'none',
  };

  return (
    <StepContainer
      title={t('hrSetup.companyProfile.title')}
      description={t('hrSetup.companyProfile.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isFirst={isFirst}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <div>
          <label
            htmlFor="hr-company-name"
            className="mb-1 block text-sm font-medium"
            style={{ color: 'var(--hr-text)' }}
          >
            {t('hrSetup.companyProfile.name')}
          </label>
          <input
            type="text"
            id="hr-company-name"
            value={sharedName || form.name}
            readOnly={Boolean(sharedName)}
            onChange={(e) => {
              setForm((f) => ({ ...f, name: e.target.value }));
              setErrors((prev) => ({ ...prev, name: '' }));
            }}
            style={inputStyle}
          />
          {sharedName && (
            <p className="mt-1 text-xs" style={{ color: 'var(--hr-text-muted)' }}>
              {t('hrSetup.companyProfile.sharedName')}
            </p>
          )}
          {errors.name && <p className="mt-1 text-xs text-red-500">{errors.name}</p>}
        </div>

        <div>
          <label
            htmlFor="hr-company-timezone"
            className="mb-1 block text-sm font-medium"
            style={{ color: 'var(--hr-text)' }}
          >
            {t('hrSetup.companyProfile.timezone')}
          </label>
          <select
            id="hr-company-timezone"
            value={form.timezone}
            onChange={(e) => {
              setForm((f) => ({ ...f, timezone: e.target.value }));
              setErrors((prev) => ({ ...prev, timezone: '' }));
            }}
            style={inputStyle}
          >
            <option value="">-</option>
            {TIMEZONES.map((tz) => (
              <option key={tz.value} value={tz.value}>
                {tz.label}
              </option>
            ))}
          </select>
          {errors.timezone && <p className="mt-1 text-xs text-red-500">{errors.timezone}</p>}
        </div>

        <div>
          <label
            htmlFor="hr-company-currency"
            className="mb-1 block text-sm font-medium"
            style={{ color: 'var(--hr-text)' }}
          >
            {t('hrSetup.companyProfile.currency')}
          </label>
          <select
            id="hr-company-currency"
            value={form.currency}
            onChange={(e) => {
              setForm((f) => ({ ...f, currency: e.target.value }));
              setErrors((prev) => ({ ...prev, currency: '' }));
            }}
            style={inputStyle}
          >
            {SETUP_CURRENCIES.map((code) => (
              <option key={code} value={code}>
                {code}, {t(`hrCommon.currencies.${code}`)}
              </option>
            ))}
          </select>
          {errors.currency && <p className="mt-1 text-xs text-red-500">{errors.currency}</p>}
        </div>
      </div>
    </StepContainer>
  );
};

export default CompanyProfileStep;
