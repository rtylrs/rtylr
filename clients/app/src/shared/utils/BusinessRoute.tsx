import { Outlet, useRouterState } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { resolveAuthUrl } from '@voxire/config';
import { parseEffectiveAccess } from '@voxire/auth';
import CreateBusinessOnboarding from '../../features/onboarding/CreateBusinessOnboarding';
import { useUserStore, type StoredUser } from '../stores/userStore';
import { normalizeConfigData, useConfigStore } from '../stores/configStore';
import { configClient, userClient } from '../lib/clients';
import { useBusinessFavicon } from '../hooks/useBusinessFavicon';
import { canAccessPortalPath } from './canAccessPortalPath';
import PortalShell from '../../shell/PortalShell';
import { Link } from '@tanstack/react-router';

const getCookieToken = (): string | null => {
  const raw = document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.slice('token='.length);
  if (!raw) return null;
  try {
    return decodeURIComponent(raw);
  } catch {
    return raw;
  }
};

function MFAAccessGate({ enrollmentRequired }: { enrollmentRequired: boolean }): React.JSX.Element {
  const { t } = useTranslation('common');

  return (
    <div
      className="flex min-h-screen items-center justify-center p-6"
      style={{ background: 'var(--dash-bg)' }}
    >
      <main
        className="w-full max-w-md rounded-2xl border p-6 text-center shadow-sm"
        style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
      >
        <h1 className="text-xl font-semibold" style={{ color: 'var(--dash-text)' }}>
          {enrollmentRequired
            ? t('common.mfaGate.enrollmentTitle')
            : t('common.mfaGate.verificationTitle')}
        </h1>
        <p className="mt-2 text-sm" style={{ color: 'var(--dash-text-muted)' }}>
          {enrollmentRequired
            ? t('common.mfaGate.enrollmentDescription')
            : t('common.mfaGate.verificationDescription')}
        </p>
        <a
          href="/security"
          className="mt-5 inline-flex min-h-11 items-center justify-center rounded-lg px-4 text-sm font-semibold"
          style={{ background: 'var(--accent)', color: 'var(--dash-on-accent)' }}
        >
          {enrollmentRequired
            ? t('common.mfaGate.openSetup')
            : t('common.mfaGate.openVerification')}
        </a>
      </main>
    </div>
  );
}

function BusinessRoute(): React.JSX.Element | null {
  const { t } = useTranslation();
  const token = getCookieToken();
  const { user, setUser, setToken } = useUserStore();
  const { setWebsiteData, setContactInfo, setMaintanance, setAboutContent } = useConfigStore();
  const logoUrl = useConfigStore((s) => s.websiteData?.logoUrl as string | undefined);
  const [loading, setLoading] = useState(true);
  const [authLoadFailed, setAuthLoadFailed] = useState(false);
  const [reloadCount, setReloadCount] = useState(0);
  const pathname = useRouterState({ select: (s) => s.location.pathname });

  useBusinessFavicon(logoUrl);

  useEffect((): void => {
    if (!token) {
      const login = new URL(resolveAuthUrl());
      login.searchParams.set('redirect', 'dash');
      login.searchParams.set(
        'returnTo',
        window.location.pathname + window.location.search + window.location.hash
      );
      window.location.href = login.href;
      return;
    }

    setToken(token);
    setLoading(true);
    setAuthLoadFailed(false);
    let leaving = false;

    userClient
      .getMe({})
      .then((res) => {
        // Nothing past /users/me answers an account whose email is
        // unverified, so it goes back to the sign-in app's verification screen.
        if (res.provisioningStatus?.emailVerified === false) {
          leaving = true;
          const verify = new URL('/verify', resolveAuthUrl());
          if (res.email) verify.searchParams.set('email', res.email);
          verify.searchParams.set('redirect', 'dash');
          verify.searchParams.set(
            'returnTo',
            window.location.pathname + window.location.search + window.location.hash
          );
          window.location.href = verify.href;
          return;
        }
        const effectiveAccess = parseEffectiveAccess(res.effectiveAccess);
        if (Number(res.businessId ?? 0) > 0 && !effectiveAccess) {
          throw new Error('Account authorization is unavailable');
        }
        setUser({
          ...res,
          effectiveAccess: effectiveAccess ?? undefined,
          provisioningStatus: res.provisioningStatus
            ? {
                mode: res.provisioningStatus.mode ?? 'invitation',
                state: res.provisioningStatus.state ?? 'pending_activation',
                emailVerified: res.provisioningStatus.emailVerified ?? false,
                passwordChangeRequired: res.provisioningStatus.passwordChangeRequired ?? false,
                mfaRequired: res.provisioningStatus.mfaRequired ?? false,
                mfaEnrolled: res.provisioningStatus.mfaEnrolled ?? false,
                mfaVerified: res.provisioningStatus.mfaVerified ?? false,
                recoveryAvailable: res.provisioningStatus.recoveryAvailable ?? false,
              }
            : undefined,
        } as StoredUser);
        configClient
          .getConfig({})
          .then((configRes) => {
            const business = configRes.business ?? {};
            const config = normalizeConfigData({
              site_name: business.name,
              logo_url: business.logoUrl,
              contact_email: business.email,
              contact_number: business.phone,
              address: business.address,
              slug: business.slug,
              website: business.website,
              country: business.country,
            });
            setWebsiteData(config);
            setContactInfo({
              contactEmail: (config.contactEmail as string | null) ?? null,
              contactNumber: (config.contactNumber as string | null) ?? null,
            });
            setMaintanance(configRes.maintenance ?? false);
            setAboutContent({ aboutContent: config.aboutContent ?? null });
          })
          .catch(() => {
            // config is non-critical - missing config won't block the app
          });
      })
      .catch(() => {
        setAuthLoadFailed(true);
      })
      .finally(() => {
        if (!leaving) setLoading(false);
      });
  }, [
    token,
    reloadCount,
    setAboutContent,
    setContactInfo,
    setMaintanance,
    setToken,
    setUser,
    setWebsiteData,
  ]);

  if (!token) return null;

  if (loading) {
    return (
      <div
        style={{
          height: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'var(--dash-bg, #fff)',
        }}
      >
        <div
          style={{
            width: 24,
            height: 24,
            border: '2px solid #e5e5e5',
            borderTopColor: '#4583C4',
            borderRadius: '50%',
            animation: 'spin 0.6s linear infinite',
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg) } }`}</style>
      </div>
    );
  }

  if (authLoadFailed || !user) {
    return (
      <main role="alert" style={{ maxWidth: 480, margin: '15vh auto', padding: 24 }}>
        <h1>{t('common.accountLoad.title')}</h1>
        <p>{t('common.accountLoad.description')}</p>
        <button type="button" onClick={() => setReloadCount((count) => count + 1)}>
          {t('common.refresh')}
        </button>
      </main>
    );
  }

  // Freshly-registered owners have no business yet (registration no longer
  // auto-creates one). Gate them into onboarding to create their first business
  // before anything else loads.
  // businessId arrives over the wire as an int64-string (e.g. "0"), so it must
  // be parsed with Number() before a truthiness check - a bare cast leaves it
  // a non-empty string, and `!"0"` is false.
  const currentBusinessId = Number(user?.businessId ?? 0);
  if (user && !currentBusinessId) {
    return <CreateBusinessOnboarding />;
  }

  const provisioningStatus = user?.provisioningStatus;
  if (
    user &&
    provisioningStatus?.mfaRequired &&
    !provisioningStatus.mfaVerified &&
    pathname !== '/security'
  ) {
    return <MFAAccessGate enrollmentRequired={!provisioningStatus.mfaEnrolled} />;
  }

  return (
    <PortalShell>
      {canAccessPortalPath(user, pathname) ? (
        <Outlet />
      ) : (
        <section role="alert" className="portal-access-denied">
          <h1>{t('common.portal.accessDeniedTitle')}</h1>
          <p>{t('common.portal.accessDeniedDescription')}</p>
          <Link to="/">{t('common.portal.backHome')}</Link>
        </section>
      )}
    </PortalShell>
  );
}

export { BusinessRoute };
