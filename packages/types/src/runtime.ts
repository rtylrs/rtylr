import type { AppId } from './user';

/**
 * Deployment mode reported by the backend `/runtime-config` endpoint.
 *
 * - `hosted`      — the SaaS product on *.rtylr.com. Behaves exactly as before;
 *   the frontend treats runtime config as a no-op overlay.
 * - `self_hosted` — a customer-operated install behind their own domain, where
 *   the prebuilt images must learn their URLs / feature posture at boot.
 */
export type DeploymentMode = 'hosted' | 'self_hosted';

/**
 * The set of absolute service/app URLs the backend advertises. Every value is
 * an absolute URL (e.g. `https://dash.example.com`). `api` is the API base and
 * `upload` is the CDN/upload base.
 */
export interface RuntimeUrls {
  auth?: string;
  dash?: string;
  pos?: string;
  erp?: string;
  hr?: string;
  crm?: string;
  finance?: string;
  flow?: string;
  insights?: string;
  recruit?: string;
  menu?: string;
  api?: string;
  upload?: string;
}

export interface RuntimeApps {
  /** App ids that are enabled for this deployment. */
  enabled?: AppId[];
  /** Map of app id → app ids it depends on (e.g. `{ recruit: ['hr'] }`). */
  dependencies?: Partial<Record<AppId, AppId[]>>;
}

/**
 * Public, non-secret license state surfaced inside `/runtime-config`. The
 * license *key* is never exposed to the browser — only this state and the
 * existing `/license/state` endpoint are consumed.
 */
export interface RuntimeLicense {
  valid?: boolean;
  status?: string;
  [key: string]: unknown;
}

/**
 * The anonymous `GET {API_URL}/runtime-config` payload (also served at
 * `/.well-known/rtylr.json`). Consumed once at app boot.
 */
export interface RuntimeConfig {
  deployment_mode?: DeploymentMode;
  self_hosted?: boolean;
  version?: string;
  urls?: RuntimeUrls;
  apps?: RuntimeApps;
  all_features_enabled?: boolean;
  license?: RuntimeLicense;
}
