import React from 'react';
import { useTranslation } from 'react-i18next';
import { RenniliLogo } from '../../shared/components/RenniliLogo';
import { VenueLogo } from '../../shared/components/VenueLogo';
import { VenueQr } from './VenueQr';

const POSTER_LANGUAGES = ['en', 'ar'] as const;
type PosterLanguage = (typeof POSTER_LANGUAGES)[number];

function PosterSteps({ lng }: { lng: PosterLanguage }): React.JSX.Element {
  const { t } = useTranslation('pager', { lng });
  const steps = [
    ['scan', t('pager.poster.stepScan')],
    ['show', t('pager.poster.stepShow')],
    ['wait', t('pager.poster.stepWait')],
  ] as const;
  return (
    <section lang={lng} dir={lng === 'ar' ? 'rtl' : 'ltr'} className="pager-poster__lang">
      <h2>{t('pager.poster.callToAction')}</h2>
      <ol>
        {steps.map(([key, step], index) => (
          <li key={key}>
            <span className="pager-poster__step" aria-hidden="true">
              {index + 1}
            </span>
            <span>{step}</span>
          </li>
        ))}
      </ol>
    </section>
  );
}

interface CounterPosterProps {
  venueName: string;
  logoUrl?: string;
  joinUrl: string;
  qrSize: number;
}

/** The counter poster: English and Arabic, whatever language the Portal is in. */
export function CounterPoster({
  venueName,
  logoUrl,
  joinUrl,
  qrSize,
}: CounterPosterProps): React.JSX.Element {
  return (
    <article className="pager-poster">
      <VenueLogo src={logoUrl} className="pager-poster__logo" />
      {venueName && <p className="pager-poster__venue">{venueName}</p>}
      <div className="pager-poster__qr">
        <VenueQr url={joinUrl} size={qrSize} />
      </div>
      <p className="pager-poster__url" dir="ltr">
        {joinUrl}
      </p>
      <div className="pager-poster__langs">
        {POSTER_LANGUAGES.map((lng) => (
          <PosterSteps key={lng} lng={lng} />
        ))}
      </div>
      <div className="pager-poster__brand">
        <RenniliLogo size={22} />
      </div>
    </article>
  );
}
