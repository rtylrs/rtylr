import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { HelmetContainer } from '@voxire/components';
import { ArrowRight, Check, LayoutGrid, RotateCcw, Smartphone } from 'lucide-react';
import Board from '../Staff/Board';
import Scan from './Scan';
import Ticket from './Ticket';
import LanguageSwitch from '../../shared/components/LanguageSwitch';
import { ByRtylr, RenniliLogo } from '../../shared/components/RenniliLogo';
import { PAGER_LANDING_PATH, pagerClaimUrl } from '../../pagerPaths';
import { PagerClientsProvider } from '../../shared/lib/pagerClients';
import { createSandboxPager } from '../../shared/lib/sandboxPager';
import { LandingFooter } from './landing/LandingFooter';
import './Try.css';

const SANDBOX_SLUG = 'demo';

/**
 * The public demo: the real staff board and the real customer screens side by
 * side, talking to an in-memory pager instead of the API. Nothing is stored,
 * and a reload or "Start over" resets the session.
 */
const Try = (): React.JSX.Element => {
  const { t, i18n } = useTranslation('pager');
  const [sessionId, setSessionId] = useState(0);
  const [sandbox, setSandbox] = useState(() =>
    createSandboxPager({ venueName: () => i18n.t('pager.try.venueName', { ns: 'pager' }) })
  );
  // Its own cache, so the demo's tickets never mix with a signed-in board's.
  const [queryClient, setQueryClient] = useState(
    () => new QueryClient({ defaultOptions: { queries: { retry: false } } })
  );
  const [phoneToken, setPhoneToken] = useState<string | null>(null);
  // Each "scan with another phone" mounts a fresh scan screen.
  const [phone, setPhone] = useState(0);
  const [hasLabelled, setHasLabelled] = useState(false);
  const [hasMarkedReady, setHasMarkedReady] = useState(false);
  const [mobileView, setMobileView] = useState<'phone' | 'board'>('phone');

  // Observe label/ready mutations locally to drive the live 3-step checklist.
  const trackedSandbox = useMemo(
    () => ({
      ...sandbox,
      labelTicket: async (
        args: Parameters<typeof sandbox.labelTicket>[0]
      ): ReturnType<typeof sandbox.labelTicket> => {
        const res = await sandbox.labelTicket(args);
        setHasLabelled(true);
        return res;
      },
      markReady: async (
        args: Parameters<typeof sandbox.markReady>[0]
      ): ReturnType<typeof sandbox.markReady> => {
        const res = await sandbox.markReady(args);
        setHasMarkedReady(true);
        return res;
      },
    }),
    [sandbox]
  );

  const clients = useMemo(
    () => ({ publicClient: trackedSandbox, staffClient: trackedSandbox, sandbox: true }),
    [trackedSandbox]
  );

  const handleStartOver = (): void => {
    setSandbox(
      createSandboxPager({ venueName: () => i18n.t('pager.try.venueName', { ns: 'pager' }) })
    );
    setQueryClient(new QueryClient({ defaultOptions: { queries: { retry: false } } }));
    setPhoneToken(null);
    setPhone((n) => n + 1);
    setHasLabelled(false);
    setHasMarkedReady(false);
    setMobileView('phone');
    setSessionId((n) => n + 1);
  };

  const steps = [
    {
      text: t('pager.try.stepPhone'),
      done: phoneToken !== null,
      active: phoneToken === null,
    },
    {
      text: t('pager.try.stepBoard'),
      done: hasLabelled,
      active: phoneToken !== null && !hasLabelled,
    },
    {
      text: t('pager.try.stepRing'),
      done: hasMarkedReady,
      active: hasLabelled && !hasMarkedReady,
    },
  ];

  return (
    <QueryClientProvider client={queryClient}>
      <PagerClientsProvider value={clients}>
        <div className="rennili-landing pager-try-shell">
          <header className="pager-try-nav">
            <div className="pager-try-nav-inner">
              <div className="rl-nav-brand">
                <a
                  href={PAGER_LANDING_PATH}
                  className="rl-nav-logo-link"
                  aria-label={t('pager.try.backToLanding')}
                >
                  <RenniliLogo size={26} />
                </a>
                <span className="rl-nav-by">
                  <ByRtylr />
                </span>
              </div>

              <div className="rl-nav-end">
                <LanguageSwitch />
                <button
                  type="button"
                  className="rl-btn rl-btn--secondary rl-btn--nav"
                  onClick={handleStartOver}
                >
                  <RotateCcw size={15} strokeWidth={2.2} aria-hidden="true" />
                  <span>{t('pager.try.startOver')}</span>
                </button>
              </div>
            </div>
          </header>

          <main className="pager-try">
            <div className="pager-try-intro-block">
              <div className="pager-try-header">
                <h1 className="pager-try-title">{t('pager.try.heading')}</h1>
                <p className="pager-try-lead">{t('pager.try.intro')}</p>
              </div>

              <ol className="pager-try-steps">
                {steps.map((step, index) => (
                  <li
                    key={step.text}
                    className={[
                      'pager-try-step',
                      step.done ? 'pager-try-step--done' : '',
                      step.active ? 'pager-try-step--active' : '',
                    ]
                      .filter(Boolean)
                      .join(' ')}
                  >
                    <span className="pager-try-step-num" aria-hidden="true">
                      {step.done ? <Check size={14} strokeWidth={3} /> : index + 1}
                    </span>
                    <span>{step.text}</span>
                  </li>
                ))}
              </ol>

              <p className="pager-try-note">{t('pager.try.sandboxNotice')}</p>
            </div>

            <div className="pager-try-stage">
              {/* Mobile segmented switcher (<940px); both panels stay mounted for desktop & accessibility. */}
              <div
                className="pager-try-mobile-tabs"
                role="group"
                aria-label={t('pager.try.mobileSwitcherLabel')}
              >
                <button
                  type="button"
                  aria-pressed={mobileView === 'phone'}
                  onClick={() => setMobileView('phone')}
                  className={`pager-try-mobile-tab ${mobileView === 'phone' ? 'pager-try-mobile-tab--active' : ''}`}
                >
                  <Smartphone size={15} strokeWidth={2.2} aria-hidden="true" />
                  <span>{t('pager.try.phoneHeading')}</span>
                  {hasMarkedReady && <span className="pager-try-tab-pulse" aria-hidden="true" />}
                </button>
                <button
                  type="button"
                  aria-pressed={mobileView === 'board'}
                  onClick={() => setMobileView('board')}
                  className={`pager-try-mobile-tab ${mobileView === 'board' ? 'pager-try-mobile-tab--active' : ''}`}
                >
                  <LayoutGrid size={15} strokeWidth={2.2} aria-hidden="true" />
                  <span>{t('pager.try.counterHeading')}</span>
                </button>
              </div>

              <div className="pager-try-grid" data-mobile-view={mobileView}>
                <section
                  className="pager-try-phone-column"
                  aria-labelledby="pager-try-phone-heading"
                >
                  <h2 id="pager-try-phone-heading" className="pager-try-label">
                    {t('pager.try.phoneHeading')}
                  </h2>
                  <div className="pager-try-phone">
                    <div className="pager-try-phone-notch" aria-hidden="true" />
                    <div className="pager-try-phone-viewport">
                      {phoneToken ? (
                        <Ticket key={`${sessionId}-${phoneToken}`} token={phoneToken} />
                      ) : (
                        <Scan
                          key={`${sessionId}-${phone}`}
                          slug={SANDBOX_SLUG}
                          onScanned={setPhoneToken}
                        />
                      )}
                    </div>
                  </div>
                  {phoneToken && (
                    <button
                      type="button"
                      className="rl-btn rl-btn--secondary rl-btn--full"
                      onClick={() => {
                        setPhoneToken(null);
                        setPhone((n) => n + 1);
                      }}
                    >
                      <Smartphone size={17} strokeWidth={2.1} aria-hidden="true" />
                      <span>{t('pager.try.anotherPhone')}</span>
                    </button>
                  )}
                </section>

                <section className="pager-try-counter" aria-labelledby="pager-try-counter-heading">
                  <h2 id="pager-try-counter-heading" className="pager-try-label">
                    {t('pager.try.counterHeading')}
                  </h2>
                  <div className="pager-try-board">
                    <Board key={sessionId} />
                  </div>
                  {hasMarkedReady && (
                    <p className="pager-try-done" role="status">
                      <Check size={16} strokeWidth={2.8} aria-hidden="true" />
                      <span>{t('pager.try.doneBanner')}</span>
                    </p>
                  )}
                </section>
              </div>
            </div>

            <section className="pager-try-claim" aria-labelledby="pager-try-claim-heading">
              <div className="pager-try-claim-copy">
                <h2 id="pager-try-claim-heading" className="pager-try-claim-title">
                  {t('pager.try.claimHeading')}
                </h2>
                <p className="pager-try-claim-body">{t('pager.try.claimBody')}</p>
              </div>
              <div className="pager-try-claim-actions">
                <a className="rl-btn rl-btn--on-forest rl-btn--lg" href={pagerClaimUrl('register')}>
                  {t('pager.try.claim')}
                  <ArrowRight
                    size={18}
                    strokeWidth={2.4}
                    aria-hidden="true"
                    className="rtl:rotate-180"
                  />
                </a>
                <p className="pager-try-signin">
                  {t('pager.try.haveAccount')}{' '}
                  <a className="pager-try-link" href={pagerClaimUrl('signIn')}>
                    {t('pager.try.signIn')}
                  </a>
                </p>
              </div>
            </section>
          </main>

          <LandingFooter onLanding={false} />
        </div>
        {/* After the screens above, so the demo's title is the one that sticks. */}
        <HelmetContainer title={t('pager.try.pageTitle')} />
      </PagerClientsProvider>
    </QueryClientProvider>
  );
};

export default Try;
