export { insightsKpisSchema, insightsDashboardSummarySchema } from './dashboard';
export type { InsightsKpis, InsightsDashboardSummary } from './dashboard';

export { revenueDataSchema, revenueByAppSchema, revenueComparisonSchema } from './revenue';
export type { RevenueData, RevenueByApp, RevenueComparison } from './revenue';

export {
  customerAnalyticsSchema,
  topCustomerSchema,
  customerChurnSchema,
  customerLtvSchema,
} from './customers';
export type { CustomerAnalytics, TopCustomer, CustomerChurn, CustomerLtv } from './customers';

export {
  productPerformanceSchema,
  productRankingSchema,
  marginAnalysisSchema,
  categoryBreakdownSchema,
} from './products';
export type {
  ProductPerformance,
  ProductRanking,
  MarginAnalysis,
  CategoryBreakdown,
} from './products';

export {
  operationsSummarySchema,
  staffUtilizationSchema,
  shiftCoverageSchema,
  timeOffTrendSchema,
  deptHeadcountSchema,
  hrDashboardSchema,
} from './operations';
export type {
  OperationsSummary,
  StaffUtilization,
  ShiftCoverage,
  TimeOffTrend,
  DeptHeadcount,
  HrDashboard,
} from './operations';

export { reportTemplateSchema } from './reports';
export type { ReportTemplate } from './reports';

export { insightsSettingsSchema, scheduledReportSchema } from './settings';
export type { InsightsSettings, ScheduledReport } from './settings';
