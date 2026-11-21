import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { resolveAppUrl } from '@voxire/config';
import { BusinessRoute, type StandaloneSurface } from '../../../../shared/utils/BusinessRoute';
import { PAGER_BOARD_PATH, portalPathForStaffPage } from '../../pagerPaths';
import { ByRtylr, RenniliLogo } from './RenniliLogo';

function StaffShell({ page }: { page: React.ReactNode }): React.JSX.Element {
  const { t } = useTranslation('pager');
  return (
    <div className="rennili-staff">
      <header className="rennili-staff-header">
        <Link to={PAGER_BOARD_PATH} className="rennili-logo" aria-label={t('pager.brand.name')}>
          <RenniliLogo size={26} />
        </Link>
        <ByRtylr />
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
};

/** The Portal's account and access checks around Rennili's staff pages. */
const RenniliStaffGate = (): React.JSX.Element | null => (
  <BusinessRoute standalone={STAFF_SURFACE} />
);

export default RenniliStaffGate;
