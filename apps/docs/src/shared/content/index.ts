import { introduction } from './introduction';
import { quickStart } from './quick-start';
import { apiOverview } from './api-overview';

export interface DocContent {
  title: string;
  body: string;
}

export const content: Record<string, DocContent> = {
  introduction,
  'quick-start': quickStart,
  'api-overview': apiOverview,
};

export { introduction, quickStart, apiOverview };
