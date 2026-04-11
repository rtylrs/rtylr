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
};
