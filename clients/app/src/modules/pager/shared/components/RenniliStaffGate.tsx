import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { resolveAppUrl } from '@voxire/config';
import { BusinessRoute, type StandaloneSurface } from '../../../../shared/utils/BusinessRoute';
import { PAGER_BOARD_PATH, pagerSignedOutUrl, portalPathForStaffPage } from '../../pagerPaths';
import { usePagerVenue } from '../lib/usePagerVenue';
import { ByRtylr, RenniliLogo } from './RenniliLogo';
import { VenueLogo } from './VenueLogo';

/** Whose board this is: the business's logo and name. */
function StaffVenue(): React.JSX.Element | null {
  const { name, logoUrl } = usePagerVenue();
  if (!name && !logoUrl) return null;
  return (
    <span className="rennili-staff-venue">
      <VenueLogo src={logoUrl} />
      {name && <span className="rennili-staff-venue__name">{name}</span>}
    </span>
  );
}

function StaffShell({ page }: { page: React.ReactNode }): React.JSX.Element {
  const { t } = useTranslation('pager');
  return (
    <div className="rennili-staff">
      <header className="rennili-staff-header">
        <Link to={PAGER_BOARD_PATH} className="rennili-logo" aria-label={t('pager.brand.name')}>
          <RenniliLogo size={26} />
        </Link>
        <ByRtylr />
        <StaffVenue />
        <a className="rennili-staff-portal" href={resolveAppUrl('dash')}>
          {t('pager.staff.portal')}
        </a>
      </header>
      <main className="rennili-staff-main">{page}</main>
    </div>
  );
}

// Module level, so the surface keeps one identity across renders.
const STAFF_SURFACE: StandaloneSurface = {
  portalPath: portalPathForStaffPage,
  shell: (page) => <StaffShell page={page} />,
  signedOutUrl: pagerSignedOutUrl,
};

/** The Portal's account and access checks around Rennili's staff pages. */
const RenniliStaffGate = (): React.JSX.Element | null => (
  <BusinessRoute standalone={STAFF_SURFACE} />
);

export default RenniliStaffGate;
