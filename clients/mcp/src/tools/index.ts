export { flowTools } from './flow/index.js';
export { crmTools } from './crm/index.js';
export { financeTools } from './finance/index.js';
export { erpTools } from './erp/index.js';
export { posTools } from './pos/index.js';
export { authTools } from './auth-tools.js';

import { flowTools } from './flow/index.js';
import { crmTools } from './crm/index.js';
import { financeTools } from './finance/index.js';
import { erpTools } from './erp/index.js';
import { posTools } from './pos/index.js';
import { authTools } from './auth-tools.js';

export const ALL_TOOLS = [...flowTools, ...crmTools, ...financeTools, ...erpTools, ...posTools, ...authTools];
