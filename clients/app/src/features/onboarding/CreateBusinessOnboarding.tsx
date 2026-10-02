import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Building2, Loader2, Check, X } from 'lucide-react';
import { setCookie, removeCookie, getDefaultCookieDomain } from '@voxire/auth';
import { useBusinessExists, useOnboardBusiness } from '../../shared/hooks/queries/useBusinesses';

// Shown by BusinessRoute when a freshly-registered owner has no business yet.
// Registration no longer auto-creates a business; the owner names it here.
// Names are globally unique, validated live via /get-business-exists.
const CreateBusinessOnboarding: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [name, setName] = useState('');
  const [debounced, setDebounced] = useState('');
  const onboard = useOnboardBusiness();

  useEffect((): (() => void) => {
    const id = setTimeout(() => setDebounced(name.trim()), 400);
    return () => clearTimeout(id);
  }, [name]);

  const { data: existsData, isFetching } = useBusinessExists(debounced);
  const longEnough = debounced.length > 1;
  const taken = longEnough && !isFetching && existsData?.exists === true;
  const available = longEnough && !isFetching && existsData?.exists === false;
  const canSubmit = name.trim().length > 1 && !taken && !isFetching && !onboard.isPending;

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();
    if (!canSubmit) return;
    onboard.mutate(name.trim(), {
      onSuccess: ({ token }) => {
        if (!token) {
          toast.error(t('onboarding.createError'));
          return;
        }
        removeCookie('token', { domain: getDefaultCookieDomain() });
        removeCookie('token');
        setCookie('token', token, {
          days: 30,
          sameSite: 'lax',
          domain: getDefaultCookieDomain(),
        });
        window.location.reload();
      },
      onError: () => toast.error(t('onboarding.createError')),
    });
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '1.5rem',
        background: 'var(--dash-bg)',
      }}
    >
      <div
        className="w-full max-w-md rounded-2xl p-8"
        style={{ background: 'var(--dash-surface)', border: '1px solid var(--dash-border)' }}
      >
        <span
          className="mb-5 inline-flex h-11 w-11 items-center justify-center rounded-xl"
          style={{ background: 'rgba(69,131,196,0.12)' }}
        >
          <Building2 size={20} style={{ color: '#4583C4' }} />
        </span>
        <h1 className="text-xl font-semibold" style={{ color: 'var(--dash-text)' }}>
          {t('onboarding.title')}
        </h1>
        <p className="mt-1.5 text-sm" style={{ color: 'var(--dash-text-muted)' }}>
          {t('onboarding.subtitle')}
        </p>

        <form onSubmit={handleSubmit} className="mt-6">
          <label
            htmlFor="business-name"
            className="mb-1.5 block text-xs font-medium"
            style={{ color: 'var(--dash-text-muted)' }}
          >
            {t('onboarding.businessName')}
          </label>
          <div className="relative">
            <input
              id="business-name"
              autoFocus
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('onboarding.businessNamePlaceholder')}
              className="w-full rounded-lg px-3 py-2.5 pr-10 outline-none"
              style={{
                fontSize: '16px',
                background: 'var(--dash-hover)',
                color: 'var(--dash-text)',
                border: `1px solid ${taken ? 'var(--dash-error, #ef4444)' : 'var(--dash-border)'}`,
              }}
              disabled={onboard.isPending}
              aria-invalid={taken}
              aria-describedby="business-name-status"
            />
            <span className="absolute right-3 top-1/2 -translate-y-1/2" aria-hidden>
              {isFetching && longEnough && (
                <Loader2
                  size={15}
                  className="animate-spin"
                  style={{ color: 'var(--dash-text-dim)' }}
                />
              )}
              {available && <Check size={15} style={{ color: '#16a34a' }} />}
              {taken && <X size={15} style={{ color: 'var(--dash-error, #ef4444)' }} />}
            </span>
          </div>
          <p
            id="business-name-status"
            aria-live="polite"
            className="mt-1.5 min-h-[1rem] text-xs"
            style={{ color: taken ? 'var(--dash-error, #ef4444)' : 'var(--dash-text-dim)' }}
          >
            {taken ? t('onboarding.nameTaken') : available ? t('onboarding.nameAvailable') : ''}
          </p>

          <button
            type="submit"
            disabled={!canSubmit}
            className="mt-5 flex w-full items-center justify-center gap-2 rounded-lg py-2.5 text-sm font-medium transition-colors disabled:opacity-50"
            style={{
              background: 'var(--accent)',
              color: 'var(--dash-on-accent)',
              minHeight: '44px',
            }}
          >
            {onboard.isPending && <Loader2 size={15} className="animate-spin" />}
            {onboard.isPending ? t('onboarding.creating') : t('onboarding.createButton')}
          </button>
        </form>
      </div>
    </div>
  );
};

export default CreateBusinessOnboarding;
