export { flowTools } from './flow/index.js';
export { crmTools } from './crm/index.js';
export { financeTools } from './finance/index.js';
export { erpTools } from './erp/index.js';

import { flowTools } from './flow/index.js';
import { crmTools } from './crm/index.js';
import { financeTools } from './finance/index.js';
import { erpTools } from './erp/index.js';

export const ALL_TOOLS = [...flowTools, ...crmTools, ...financeTools, ...erpTools];
