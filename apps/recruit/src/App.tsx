import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { AppGuard } from '@voxire/auth';
import { CookieBanner } from '@voxire/consent';

import api from './shared/lib/api';
import RootLanding from './features/public/RootLanding';
import CareersIndex from './features/public/CareersIndex';
import CareerDetail from './features/public/CareerDetail';
import ApplyPage from './features/public/ApplyPage';
import NotFound from './features/public/NotFound';
import AdminLayout from './features/admin/AdminLayout';
import CareersAdmin from './features/admin/CareersAdmin';
import SubmissionsAdmin from './features/admin/SubmissionsAdmin';
import CustomizationAdmin from './features/admin/CustomizationAdmin';
import SectionsAdmin from './features/admin/SectionsAdmin';

/**
 * Recruit is a split app: the `/admin/*` routes are the HR back-office
 * (gated behind AppGuard), and everything else is the public careers
 * site (keyed by business slug in the URL).
 *
 * The split lives in routing rather than in separate Vite apps so a
 * business can preview their own public page while logged in without
 * bouncing across subdomains.
 */
const AdminShell: React.FC = (): React.JSX.Element => (
  <AppGuard
    appId="recruit"
    fetchUser={(): Promise<Record<string, unknown>> =>
      api.get('/user/me').then((r: { data: Record<string, unknown> }) => r.data)
    }
  >
    <Routes>
      <Route path="/" element={<AdminLayout />}>
        <Route index element={<CareersAdmin />} />
        <Route path="careers" element={<CareersAdmin />} />
        <Route path="submissions" element={<SubmissionsAdmin />} />
        <Route path="sections" element={<SectionsAdmin />} />
        <Route path="customization" element={<CustomizationAdmin />} />
      </Route>
    </Routes>
  </AppGuard>
);

const App: React.FC = (): React.JSX.Element => (
  <>
    <Routes>
      <Route path="/" element={<RootLanding />} />
      <Route path="/admin/*" element={<AdminShell />} />
      <Route path="/:slug" element={<CareersIndex />} />
      <Route path="/:slug/:careerSlug" element={<CareerDetail />} />
      <Route path="/:slug/:careerSlug/apply" element={<ApplyPage />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
    {/* Public careers pages need the strict-opt-in banner (candidate
        traffic is mostly EU); admin routes don't strictly need it but
        showing it there too is cheaper than branching on path. */}
    <CookieBanner />
  </>
);

export default App;
