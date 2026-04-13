import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
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
}

const CompanyProfileStep = ({
  onNext,
  onBack,
  isFirst,
  isSubmitting,
}: CompanyProfileStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [form, setForm] = useState<CompanyProfileFormData>({
    name: '',
    industry: '',
    country: '',
    timezone: '',
    currency: 'USD',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (): void => {
    const result = companyProfileSchema.safeParse(form);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const field = issue.path[0];
        if (typeof field === 'string') {
          fieldErrors[field] = t('common.required');
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
    fontSize: '14px',
    outline: 'none',
  };

  return (
    <StepContainer
      title={t('setup.companyProfile.title')}
      description={t('setup.companyProfile.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isFirst={isFirst}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <div>
          <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('setup.companyProfile.name')}
          </label>
          <input
            type="text"
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            style={inputStyle}
          />
          {errors.name && <p className="mt-1 text-xs text-red-500">{errors.name}</p>}
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('setup.companyProfile.industry')}
          </label>
          <select
            value={form.industry}
            onChange={(e) => setForm((f) => ({ ...f, industry: e.target.value }))}
            style={inputStyle}
          >
            <option value="">—</option>
            <option value="restaurant">Restaurant</option>
            <option value="retail">Retail</option>
            <option value="technology">Technology</option>
            <option value="healthcare">Healthcare</option>
            <option value="education">Education</option>
            <option value="legal">Legal</option>
            <option value="finance">Finance</option>
            <option value="construction">Construction</option>
            <option value="other">Other</option>
          </select>
          {errors.industry && <p className="mt-1 text-xs text-red-500">{errors.industry}</p>}
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
              {t('setup.companyProfile.country')}
            </label>
            <input
              type="text"
              value={form.country}
              onChange={(e) => setForm((f) => ({ ...f, country: e.target.value }))}
              style={inputStyle}
            />
            {errors.country && <p className="mt-1 text-xs text-red-500">{errors.country}</p>}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
              {t('setup.companyProfile.timezone')}
            </label>
            <select
              value={form.timezone}
              onChange={(e) => setForm((f) => ({ ...f, timezone: e.target.value }))}
              style={inputStyle}
            >
              <option value="">—</option>
              <option value="America/New_York">Eastern (US)</option>
              <option value="America/Chicago">Central (US)</option>
              <option value="America/Denver">Mountain (US)</option>
              <option value="America/Los_Angeles">Pacific (US)</option>
              <option value="Europe/London">London</option>
              <option value="Europe/Paris">Paris</option>
              <option value="Europe/Berlin">Berlin</option>
              <option value="Asia/Dubai">Dubai</option>
              <option value="Asia/Beirut">Beirut</option>
              <option value="Asia/Riyadh">Riyadh</option>
              <option value="Asia/Tokyo">Tokyo</option>
              <option value="Australia/Sydney">Sydney</option>
            </select>
            {errors.timezone && <p className="mt-1 text-xs text-red-500">{errors.timezone}</p>}
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('setup.companyProfile.currency')}
          </label>
          <select
            value={form.currency}
            onChange={(e) => setForm((f) => ({ ...f, currency: e.target.value }))}
            style={inputStyle}
          >
            <option value="USD">USD — US Dollar</option>
            <option value="EUR">EUR — Euro</option>
            <option value="GBP">GBP — British Pound</option>
            <option value="AED">AED — UAE Dirham</option>
            <option value="SAR">SAR — Saudi Riyal</option>
            <option value="LBP">LBP — Lebanese Pound</option>
            <option value="JPY">JPY — Japanese Yen</option>
            <option value="AUD">AUD — Australian Dollar</option>
            <option value="CAD">CAD — Canadian Dollar</option>
          </select>
          {errors.currency && <p className="mt-1 text-xs text-red-500">{errors.currency}</p>}
        </div>
      </div>
    </StepContainer>
  );
};

export default CompanyProfileStep;
