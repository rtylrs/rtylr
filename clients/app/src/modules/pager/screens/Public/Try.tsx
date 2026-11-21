import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { HelmetContainer } from '@voxire/components';
import Board from '../Staff/Board';
import Scan from './Scan';
import Ticket from './Ticket';
import LanguageSwitch from '../../shared/components/LanguageSwitch';
import { ByRtylr, RenniliLogo } from '../../shared/components/RenniliLogo';
import { pagerClaimUrl } from '../../pagerPaths';
import { PagerClientsProvider } from '../../shared/lib/pagerClients';
import { createSandboxPager } from '../../shared/lib/sandboxPager';
import './Try.css';

const SANDBOX_SLUG = 'demo';

/**
 * The public demo: the real staff board and the real customer screens side by
 * side, talking to an in-memory pager instead of the API. Nothing is stored,
 * and a reload starts over.
 */
const Try = (): React.JSX.Element => {
  const { t, i18n } = useTranslation('pager');
  const [sandbox] = useState(() =>
    createSandboxPager({ venueName: () => i18n.t('pager.try.venueName', { ns: 'pager' }) })
  );
  // Its own cache, so the demo's tickets never mix with a signed-in board's.
  const [queryClient] = useState(
    () => new QueryClient({ defaultOptions: { queries: { retry: false } } })
  );
  const clients = useMemo(
    () => ({ publicClient: sandbox, staffClient: sandbox, sandbox: true }),
    [sandbox]
  );
  const [phoneToken, setPhoneToken] = useState<string | null>(null);
  // Each "scan with another phone" mounts a fresh scan screen.
  const [phone, setPhone] = useState(0);

  return (
    <QueryClientProvider client={queryClient}>
      <PagerClientsProvider value={clients}>
        <div className="pager-try">
          <header className="pager-try-header">
            <div className="pager-topbar">
              <div className="pager-venue">
                <RenniliLogo size={26} />
                <ByRtylr />
              </div>
              <LanguageSwitch />
            </div>
            <h1 className="pager-heading">{t('pager.try.heading')}</h1>
            <p className="pager-text">{t('pager.try.intro')}</p>
            <ol className="pager-try-steps">
              <li>{t('pager.try.stepPhone')}</li>
              <li>{t('pager.try.stepBoard')}</li>
              <li>{t('pager.try.stepRing')}</li>
            </ol>
            <p className="pager-text pager-try-note">{t('pager.try.sandboxNotice')}</p>
          </header>

          <div className="pager-try-grid">
            <section className="pager-try-phone-column" aria-labelledby="pager-try-phone-heading">
              <h2 id="pager-try-phone-heading" className="pager-try-label">
                {t('pager.try.phoneHeading')}
              </h2>
              <div className="pager-try-phone">
                {phoneToken ? (
                  <Ticket token={phoneToken} />
                ) : (
                  <Scan key={phone} slug={SANDBOX_SLUG} onScanned={setPhoneToken} />
                )}
              </div>
              {phoneToken && (
                <button
                  type="button"
                  className="pager-button pager-button--secondary"
                  onClick={() => {
                    setPhoneToken(null);
                    setPhone((n) => n + 1);
                  }}
                >
                  {t('pager.try.anotherPhone')}
                </button>
              )}
            </section>

            <section className="pager-try-counter" aria-labelledby="pager-try-counter-heading">
              <h2 id="pager-try-counter-heading" className="pager-try-label">
                {t('pager.try.counterHeading')}
              </h2>
              <div className="pager-try-board">
                <Board />
              </div>
            </section>
          </div>

          <section className="pager-try-claim" aria-labelledby="pager-try-claim-heading">
            <h2 id="pager-try-claim-heading" className="pager-heading">
              {t('pager.try.claimHeading')}
            </h2>
            <p className="pager-text">{t('pager.try.claimBody')}</p>
            <a className="pager-button" href={pagerClaimUrl('register')}>
              {t('pager.try.claim')}
            </a>
            <p className="pager-text">
              {t('pager.try.haveAccount')}{' '}
              <a className="pager-try-link" href={pagerClaimUrl('signIn')}>
                {t('pager.try.signIn')}
              </a>
            </p>
          </section>
        </div>
        {/* After the screens above, so the demo's title is the one that sticks. */}
        <HelmetContainer title={t('pager.try.pageTitle')} />
      </PagerClientsProvider>
    </QueryClientProvider>
  );
};

export default Try;
