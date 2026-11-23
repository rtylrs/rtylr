import React, { useLayoutEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { HelmetContainer } from '@voxire/components';
import {
  ArrowRight,
  BatteryFull,
  Check,
  LayoutGrid,
  Plus,
  RotateCcw,
  Signal,
  Smartphone,
  Wifi,
} from 'lucide-react';
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

type DemoView = 'phone' | 'board';

// The customer screens are laid out at an iPhone's size and scaled to fit the
// frame, so the demo shows the proportions a real phone would.
const SCREEN_WIDTH = 390;

const DemoPhone = ({ children }: { children: React.ReactNode }): React.JSX.Element => {
  const screen = useRef<HTMLDivElement>(null);
  const [scale, setScale] = useState(0.75);

  useLayoutEffect(() => {
    const element = screen.current;
    if (!element) return undefined;
    const fit = (): void => {
      if (element.clientWidth > 0) setScale(element.clientWidth / SCREEN_WIDTH);
    };
    fit();
    if (typeof ResizeObserver === 'undefined') return undefined;
    const observer = new ResizeObserver(fit);
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  return (
    <div className="pager-try-phone">
      <div ref={screen} className="pager-try-phone-screen">
        <div className="pager-try-phone-canvas" style={{ transform: `scale(${scale})` }}>
          <div className="pager-try-statusbar" aria-hidden="true">
            <span className="pager-try-statusbar-time">9:41</span>
            <span className="pager-try-island" />
            <span className="pager-try-statusbar-icons">
              <Signal size={17} strokeWidth={2.6} />
              <Wifi size={17} strokeWidth={2.6} />
              <BatteryFull size={22} strokeWidth={2} />
            </span>
          </div>
          <div className="pager-try-phone-viewport">{children}</div>
        </div>
      </div>
    </div>
  );
};

/**
 * The public demo: the real staff board and the real customer screens side by
 * side, talking to an in-memory pager instead of the API. Nothing is stored,
 * and a reload or "Start over" resets the session.
 */
const Try = (): React.JSX.Element => {
  const { t, i18n } = useTranslation('pager');
  const demoVenueName = t('pager.try.venueName');
  const venueNameRef = useRef(demoVenueName);
  venueNameRef.current = demoVenueName;
  const [sessionId, setSessionId] = useState(0);
  const [sandbox, setSandbox] = useState(() =>
    createSandboxPager({
      venueName: () =>
        venueNameRef.current !== 'pager.try.venueName'
          ? venueNameRef.current
          : i18n.t('pager.try.venueName'),
    })
  );
  // Its own cache, so the demo's tickets never mix with a signed-in board's.
  const [queryClient, setQueryClient] = useState(
    () => new QueryClient({ defaultOptions: { queries: { retry: false } } })
  );
  const [phoneToken, setPhoneToken] = useState<string | null>(null);
  // Each "scan with another phone" mounts a fresh scan screen.
  const [phone, setPhone] = useState(0);
  const [hasMarkedReady, setHasMarkedReady] = useState(false);
  const [mobileView, setMobileView] = useState<DemoView>('phone');

  // The checklist moves on when the visitor marks an order ready on the board.
  const trackedSandbox = useMemo(
    () => ({
      ...sandbox,
      markReady: async (
        args: Parameters<typeof sandbox.markReady>[0]
      ): ReturnType<typeof sandbox.markReady> => {
        const res = await sandbox.markReady(args);
        setHasMarkedReady(true);
        // On a phone-sized screen, show the phone ringing.
        setMobileView('phone');
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
      createSandboxPager({
        venueName: () =>
          venueNameRef.current !== 'pager.try.venueName'
            ? venueNameRef.current
            : i18n.t('pager.try.venueName'),
      })
    );
    setQueryClient(new QueryClient({ defaultOptions: { queries: { retry: false } } }));
    setPhoneToken(null);
    setPhone((n) => n + 1);
    setHasMarkedReady(false);
    setMobileView('phone');
    setSessionId((n) => n + 1);
  };

  const steps = [
    {
      label: t('pager.try.stepPhoneShort'),
      text: t('pager.try.stepPhone'),
      done: phoneToken !== null,
    },
    { label: t('pager.try.stepBoardShort'), text: t('pager.try.stepBoard'), done: hasMarkedReady },
    { label: t('pager.try.stepRingShort'), text: t('pager.try.stepRing'), done: hasMarkedReady },
  ];
  const current = steps.findIndex((step) => !step.done);
  const finished = current === -1;
  // Where the next move happens, so a phone-sized screen can send the visitor there.
  const nextView: DemoView | null = finished ? null : current === 2 ? 'phone' : 'board';

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
            <header className="pager-try-hero">
              <p className="pager-try-badge">
                <span className="pager-try-badge-dot" aria-hidden="true" />
                {t('pager.try.demoBadge')}
              </p>
              <h1 className="pager-try-title">{t('pager.try.heading')}</h1>
              <p className="pager-try-lead">{t('pager.try.intro')}</p>
            </header>

            <div className="pager-try-stage">
              <div className="pager-try-guide">
                <ol className="pager-try-steps">
                  {steps.map((step, index) => (
                    <li
                      key={step.label}
                      className={[
                        'pager-try-step',
                        step.done ? 'pager-try-step--done' : '',
                        index === current ? 'pager-try-step--active' : '',
                      ]
                        .filter(Boolean)
                        .join(' ')}
                      aria-current={index === current ? 'step' : undefined}
                    >
                      <span className="pager-try-step-num" aria-hidden="true">
                        {step.done ? <Check size={13} strokeWidth={3} /> : index + 1}
                      </span>
                      <span className="pager-try-step-label">{step.label}</span>
                    </li>
                  ))}
                </ol>
                <div
                  className={`pager-try-instruction ${finished ? 'pager-try-instruction--done' : ''}`}
                  role="status"
                >
                  <p>{finished ? t('pager.try.doneBanner') : steps[current].text}</p>
                  {nextView && nextView !== mobileView && (
                    <button
                      type="button"
                      className="pager-try-jump"
                      onClick={() => setMobileView(nextView)}
                    >
                      {t(nextView === 'board' ? 'pager.try.goToCounter' : 'pager.try.goToPhone')}
                      <ArrowRight
                        size={15}
                        strokeWidth={2.4}
                        aria-hidden="true"
                        className="rtl:rotate-180"
                      />
                    </button>
                  )}
                </div>
              </div>

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
                  <span>{t('pager.try.phoneTab')}</span>
                  {nextView === 'phone' && mobileView !== 'phone' && (
                    <span className="pager-try-tab-pulse" aria-hidden="true" />
                  )}
                </button>
                <button
                  type="button"
                  aria-pressed={mobileView === 'board'}
                  onClick={() => setMobileView('board')}
                  className={`pager-try-mobile-tab ${mobileView === 'board' ? 'pager-try-mobile-tab--active' : ''}`}
                >
                  <LayoutGrid size={15} strokeWidth={2.2} aria-hidden="true" />
                  <span>{t('pager.try.counterTab')}</span>
                  {nextView === 'board' && mobileView !== 'board' && (
                    <span className="pager-try-tab-pulse" aria-hidden="true" />
                  )}
                </button>
              </div>

              <div className="pager-try-grid" data-mobile-view={mobileView}>
                <section
                  className="pager-try-phone-column"
                  aria-labelledby="pager-try-phone-heading"
                >
                  <h2 id="pager-try-phone-heading" className="pager-try-label">
                    <Smartphone size={14} strokeWidth={2.4} aria-hidden="true" />
                    {t('pager.try.phoneHeading')}
                  </h2>
                  <DemoPhone>
                    {phoneToken ? (
                      <Ticket key={`${sessionId}-${phoneToken}`} token={phoneToken} />
                    ) : (
                      <Scan
                        key={`${sessionId}-${phone}`}
                        slug={SANDBOX_SLUG}
                        onScanned={setPhoneToken}
                      />
                    )}
                  </DemoPhone>
                  {phoneToken && (
                    <button
                      type="button"
                      className="pager-try-another"
                      onClick={() => {
                        setPhoneToken(null);
                        setPhone((n) => n + 1);
                      }}
                    >
                      <Plus size={15} strokeWidth={2.4} aria-hidden="true" />
                      <span>{t('pager.try.anotherPhone')}</span>
                    </button>
                  )}
                </section>

                <section className="pager-try-counter" aria-labelledby="pager-try-counter-heading">
                  <h2 id="pager-try-counter-heading" className="pager-try-label">
                    <LayoutGrid size={14} strokeWidth={2.4} aria-hidden="true" />
                    {t('pager.try.counterHeading')}
                  </h2>
                  <div className="pager-try-board">
                    <Board key={sessionId} />
                  </div>
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
