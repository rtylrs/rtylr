import { introduction } from './introduction';
import { quickStart } from './quick-start';
import { apiOverview } from './api-overview';
import { pos } from './pos';
import { erp } from './erp';
import { hr } from './hr';
import { crm } from './crm';
import { finance } from './finance';
import { flow } from './flow';
import { store } from './store';
import { book } from './book';
import { support } from './support';
import { insights } from './insights';
import { hub } from './hub';
import { addons } from './addons';
import { multiTenant } from './multi-tenant';
import { faq } from './faq';
import {
  selfHostingOverview,
  selfHostingLicense,
  selfHostingImages,
  selfHostingHelm,
  selfHostingCompose,
  selfHostingDns,
  selfHostingEnv,
  selfHostingUpgrade,
} from './self-hosting';

export interface DocContent {
  title: string;
  body: string;
}

export const content: Record<string, DocContent> = {
  introduction,
  'quick-start': quickStart,
  'api-overview': apiOverview,
  pos,
  erp,
  hr,
  crm,
  finance,
  flow,
  store,
  book,
  support,
  insights,
  hub,
  addons,
  'multi-tenant': multiTenant,
  faq,
  'self-hosting': selfHostingOverview,
  'self-hosting-license': selfHostingLicense,
  'self-hosting-images': selfHostingImages,
  'self-hosting-helm': selfHostingHelm,
  'self-hosting-compose': selfHostingCompose,
  'self-hosting-dns': selfHostingDns,
  'self-hosting-env': selfHostingEnv,
  'self-hosting-upgrade': selfHostingUpgrade,
};

export {
  introduction,
  quickStart,
  apiOverview,
  pos,
  erp,
  hr,
  crm,
  finance,
  flow,
  store,
  book,
  support,
  insights,
  hub,
  addons,
  multiTenant,
  faq,
  selfHostingOverview,
  selfHostingLicense,
  selfHostingImages,
  selfHostingHelm,
  selfHostingCompose,
  selfHostingDns,
  selfHostingEnv,
  selfHostingUpgrade,
};
