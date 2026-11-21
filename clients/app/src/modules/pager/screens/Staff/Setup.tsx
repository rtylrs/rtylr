import React, { useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Navigate, useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { Alert, Spinner } from '@voxire/ui';
import { BellRing } from 'lucide-react';
import { Button } from '../../../../shared/components/ui';
import { businessClient, configClient } from '../../../../shared/lib/clients';
import { useUserStore } from '../../../../shared/stores/userStore';
import { PAGER_BOARD_PATH, pagerJoinUrl } from '../../pagerPaths';
import { errorStatus } from '../../shared/lib/apiError';
import {
  claimPager,
  isPhoneValid,
  isSlugUsable,
  normalisePhone,
  normaliseSlug,
} from '../../shared/lib/claimPager';
import { staffPagerClient } from '../../shared/lib/clients';

/**
 * The last step of the pager sign-up, reached from the demo through the
 * ordinary registration: the venue leaves its phone number, gets a link for
 * its poster and lands on the board with the pager on.
 */
const Setup = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const phoneId = useId();
  const errorId = useId();
  const slugId = useId();
  const slugErrorId = useId();
  const user = useUserStore((state) => state.user);
  const canClaim = useUserStore(
    (state) =>
      Boolean(state.user?.effectiveAccess?.isOwner) || Boolean(state.user?.effectiveAccess?.isAdmin)
  );
  const [typedPhone, setTypedPhone] = useState<string | null>(null);
  const [showInvalid, setShowInvalid] = useState(false);
  // A setup link sent to a new venue names its slug; the owner confirms it here.
  const [linkSlug] = useState(() =>
    normaliseSlug(new URLSearchParams(window.location.search).get('slug') ?? '')
  );
  const [typedSlug, setTypedSlug] = useState<string | null>(null);
  const [showSlugInvalid, setShowSlugInvalid] = useState(false);

  const business = useQuery({
    queryKey: ['pager', 'setup', 'business'],
    queryFn: async () => (await configClient.getConfig({})).business ?? {},
  });
  const settings = useQuery({
    queryKey: ['pager', 'settings'],
    queryFn: () => staffPagerClient.getSettings({}),
  });

  const venueName = business.data?.name?.trim() || user?.businessName?.trim() || '';
  const slug = business.data?.slug?.trim() ?? '';
  const phone = normalisePhone(typedPhone ?? business.data?.phone ?? '');
  const phoneValid = isPhoneValid(phone);
  const asksSlug = !slug && linkSlug !== '';
  const chosenSlug = normaliseSlug(typedSlug ?? linkSlug);
  const slugValid = isSlugUsable(chosenSlug);

  const claim = useMutation({
    mutationFn: () =>
      claimPager(
        {
          savePhone: (value) => configClient.updateConfig({ contactNumber: value }),
          setSlug: (value) => businessClient.setSlug({ slug: value }),
          enablePager: () => staffPagerClient.updateSettings({ enabled: true }),
        },
        { phone, venueName, slug, chosenSlug: asksSlug ? chosenSlug : undefined }
      ),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['pager'] });
      void navigate({ to: PAGER_BOARD_PATH });
    },
  });

  if (business.isPending || settings.isPending) {
    return (
      <div className="flex min-h-48 items-center justify-center">
        <Spinner />
      </div>
    );
  }

  // Coming back to this page after the claim, or already set up.
  if (settings.data?.enabled && slug) return <Navigate to={PAGER_BOARD_PATH} replace />;

  if (business.isError || settings.isError) {
    return (
      <div className="space-y-2">
        <Alert variant="destructive">{t('pager.setup.loadFailed')}</Alert>
        <Button
          size="sm"
          variant="outline"
          onClick={() => {
            void business.refetch();
            void settings.refetch();
          }}
        >
          {t('pager.setup.retry')}
        </Button>
      </div>
    );
  }

  const claimStatus = claim.isError ? errorStatus(claim.error) : undefined;
  const claimError =
    claim.isError &&
    (claimStatus === 403
      ? 'adminOnly'
      : claimStatus === 409 && asksSlug
        ? 'slugTaken'
        : 'claimFailed');

  return (
    <>
      <HelmetContainer title={t('pager.setup.pageTitle')} />
      <div
        className="mx-auto w-full max-w-md rounded-2xl p-6"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      >
        <div className="flex items-center gap-2">
          <BellRing size={18} aria-hidden="true" style={{ color: 'var(--dash-text-muted)' }} />
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            {t('pager.setup.heading', { venue: venueName })}
          </h1>
        </div>
        <p className="mt-1 text-sm" style={{ color: 'var(--dash-text-dim)' }}>
          {t('pager.setup.intro')}
        </p>

        {!canClaim ? (
          <Alert variant="warning" className="mt-5">
            {t('pager.setup.adminOnly')}
          </Alert>
        ) : (
          <form
            className="mt-5 space-y-3"
            noValidate
            onSubmit={(event) => {
              event.preventDefault();
              if (!phoneValid || (asksSlug && !slugValid)) {
                setShowInvalid(!phoneValid);
                setShowSlugInvalid(asksSlug && !slugValid);
                return;
              }
              if (!claim.isPending) claim.mutate();
            }}
          >
            <label
              htmlFor={phoneId}
              className="block text-xs font-semibold"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              {t('pager.setup.phoneLabel')}
            </label>
            <input
              id={phoneId}
              type="tel"
              dir="ltr"
              inputMode="tel"
              autoComplete="tel"
              required
              value={typedPhone ?? business.data?.phone ?? ''}
              aria-invalid={showInvalid && !phoneValid}
              aria-describedby={showInvalid && !phoneValid ? errorId : undefined}
              onChange={(event) => {
                setTypedPhone(event.target.value);
                setShowInvalid(false);
              }}
              className="h-11 w-full rounded-lg px-3 outline-none focus:border-[var(--dash-border-focus)]"
              style={{
                fontSize: '16px',
                background: 'var(--dash-input-bg)',
                border: '1px solid var(--dash-border)',
                color: 'var(--dash-text)',
              }}
            />
            {showInvalid && !phoneValid && (
              <p
                id={errorId}
                role="alert"
                className="text-xs"
                style={{ color: 'var(--dash-error, #ef4444)' }}
              >
                {t('pager.setup.phoneInvalid')}
              </p>
            )}
            <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
              {t('pager.setup.phoneHint')}
            </p>
            {asksSlug && (
              <>
                <label
                  htmlFor={slugId}
                  className="block pt-2 text-xs font-semibold"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  {t('pager.setup.slugLabel')}
                </label>
                <input
                  id={slugId}
                  type="text"
                  dir="ltr"
                  autoCapitalize="none"
                  autoCorrect="off"
                  spellCheck={false}
                  required
                  value={typedSlug ?? linkSlug}
                  aria-invalid={showSlugInvalid && !slugValid}
                  aria-describedby={`${slugId}-hint${showSlugInvalid && !slugValid ? ` ${slugErrorId}` : ''}`}
                  onChange={(event) => {
                    setTypedSlug(event.target.value);
                    setShowSlugInvalid(false);
                  }}
                  className="h-11 w-full rounded-lg px-3 outline-none focus:border-[var(--dash-border-focus)]"
                  style={{
                    fontSize: '16px',
                    background: 'var(--dash-input-bg)',
                    border: '1px solid var(--dash-border)',
                    color: 'var(--dash-text)',
                  }}
                />
                {showSlugInvalid && !slugValid && (
                  <p
                    id={slugErrorId}
                    role="alert"
                    className="text-xs"
                    style={{ color: 'var(--dash-error, #ef4444)' }}
                  >
                    {t('pager.setup.slugInvalid')}
                  </p>
                )}
                <p
                  id={`${slugId}-hint`}
                  className="text-xs"
                  style={{ color: 'var(--dash-text-dim)' }}
                >
                  {t('pager.setup.slugHint', {
                    link: pagerJoinUrl(slugValid ? chosenSlug : linkSlug),
                  })}
                </p>
              </>
            )}
            {claimError && <Alert variant="destructive">{t(`pager.setup.${claimError}`)}</Alert>}
            <Button type="submit" className="h-11 w-full" disabled={claim.isPending}>
              {claim.isPending ? t('pager.setup.submitting') : t('pager.setup.submit')}
            </Button>
          </form>
        )}
      </div>
    </>
  );
};

export default Setup;
