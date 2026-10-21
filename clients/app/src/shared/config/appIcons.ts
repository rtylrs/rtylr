import {
  HomeIcon,
  InsightsIcon,
  PosIcon,
  ErpIcon,
  HrIcon,
  CrmIcon,
  FinanceIcon,
  FlowIcon,
  MenuIcon,
  RecruitIcon,
} from '@voxire/ui';

export const APP_ICONS = {
  dash: HomeIcon,
  insights: InsightsIcon,
  pos: PosIcon,
  erp: ErpIcon,
  hr: HrIcon,
  crm: CrmIcon,
  finance: FinanceIcon,
  flow: FlowIcon,
  menu: MenuIcon,
  recruit: RecruitIcon,
} as const;

export const WORKSPACE_APP_IDS = [
  'insights',
  'pos',
  'erp',
  'hr',
  'crm',
  'finance',
  'flow',
  'recruit',
] as const;
