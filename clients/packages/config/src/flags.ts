/**
 * Feature-flag keys, typed so nobody ships a typo as a flag name.
 *
 * Rule of thumb:
 *   - `app.<name>` gates a whole product surface (ERP, CRM, etc.)
 *     and is OFF in prod until that surface is wired end-to-end.
 *   - `app.<name>.<sub>` gates a feature inside an already-shipped
 *     product (Stripe payments inside POS, the apply file upload
 *     inside recruit, etc.). ON during development, OFF in prod
 *     until the sub-feature is launch-ready.
 *
 * Flags resolve in this order:
 *   1. PostHog (live decision - feature rollouts, cohort toggles)
 *   2. `VITE_APP_*_ENABLED` env var (build-time fallback, read from
 *      import.meta.env - per .env.example at repo root)
 *   3. Hard-coded default listed below.
 *
 * See `appGuards.tsx` for the React components that consume these.
 */

export const FLAG_KEYS = {
  APP_RECRUIT: 'app.recruit',
  APP_ERP: 'app.erp',
  APP_CRM: 'app.crm',
  APP_FINANCE: 'app.finance',
  APP_INSIGHTS: 'app.insights',
  APP_MENU: 'app.menu',
  APP_POS: 'app.pos',
  APP_RECRUIT_FILE_UPLOAD: 'app.recruit.fileUpload',
  /**
   * Flow's UI-only stub routes (Timeline, CalendarView, Roadmap,
   * Insights, Team, ProjectOverview/Timeline/Dashboard). Off in prod
   * until these screens render real data - see vault flow.md P0 #4.
   */
  APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes',
} as const;

export type FlagKey = (typeof FLAG_KEYS)[keyof typeof FLAG_KEYS];

/**
 * Prod defaults - what the flag evaluates to when PostHog is
 * unreachable AND the env var is unset. Everything off until proven
 * otherwise. Landing, dash, auth, hr, flow are already shipped and
 * don't have flags; they always render.
 */
export const FLAG_PROD_DEFAULTS: Record<FlagKey, boolean> = {
  [FLAG_KEYS.APP_RECRUIT]: true,
  [FLAG_KEYS.APP_ERP]: true,
  [FLAG_KEYS.APP_CRM]: true,
  [FLAG_KEYS.APP_FINANCE]: true,
  [FLAG_KEYS.APP_INSIGHTS]: true,
  [FLAG_KEYS.APP_MENU]: true,
  [FLAG_KEYS.APP_POS]: true,
  [FLAG_KEYS.APP_RECRUIT_FILE_UPLOAD]: true, // storage & presigned R2 upload wired
  [FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES]: false, // off until screens render real data
};

/**
 * Map a flag key to the `VITE_APP_*_ENABLED` env var the plan's
 * `.env.example` documents. Anything unmapped (sub-flags) isn't
 * env-overridable - only PostHog + the prod default apply.
 */
export const FLAG_ENV_OVERRIDES: Partial<Record<FlagKey, string>> = {
  [FLAG_KEYS.APP_RECRUIT]: 'VITE_APP_RECRUIT_ENABLED',
  [FLAG_KEYS.APP_ERP]: 'VITE_APP_ERP_ENABLED',
  [FLAG_KEYS.APP_CRM]: 'VITE_APP_CRM_ENABLED',
  [FLAG_KEYS.APP_FINANCE]: 'VITE_APP_FINANCE_ENABLED',
  [FLAG_KEYS.APP_INSIGHTS]: 'VITE_APP_INSIGHTS_ENABLED',
  [FLAG_KEYS.APP_MENU]: 'VITE_APP_MENU_ENABLED',
  [FLAG_KEYS.APP_POS]: 'VITE_APP_POS_ENABLED',
};
