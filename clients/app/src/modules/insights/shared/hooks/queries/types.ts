// Local type definitions for /insights/* responses. These mirror the Go
// payload structs under onion/internal/modules/insights/internal/application
// and handlers, with snake_case keys automatically converted to camelCase by
// the @voxire/api response interceptor.

export interface MoneyCard {
  value: number;
  deltaPct: number;
}

export interface CountCard {
  value: number;
  deltaPct: number;
}

export interface PctCard {
  value: number;
  deltaPct: number;
}

export interface WorkProjectProgress {
  id: string;
  name: string;
  totalTasks: number;
  completedTasks: number;
  progressPct: number;
}

export interface WorkSummaryResponse {
  totalTasks: number;
  completedTasks: number;
  overdueTasks: number;
  projects: WorkProjectProgress[];
}

// CRM deal values are in minor units of the insights default currency: deals
// carry no currency of their own.
export interface SalesSummaryResponse {
  openPipelineValue: number;
  openDealCount: number;
  won: MoneyCard;
  recentWonDealCount: number;
  // Monthly recurring revenue of won monthly and yearly deals, compared with
  // the start of this month.
  mrr: MoneyCard;
  activeContractCount: number;
}

// Time logged in Work (Flow), in minutes. Only public projects are listed;
// private ones count in the total without being named.
export interface ProjectHoursThisMonth {
  projectId: string;
  minutes: number;
}

export interface HoursSummaryResponse {
  // Minutes logged this month to date, against the same span last month.
  logged: CountCard;
  priorMinutes: number;
  projects: ProjectHoursThisMonth[];
}

// A Work project the viewer may open a project dashboard for: every public
// project and the private ones they are a member of.
export interface DashboardProjectOption {
  id: string;
  name: string;
  status: string;
  visibility: string;
}

export interface ProjectOverdueTask {
  id: string;
  title: string;
  // YYYY-MM-DD
  dueDate: string;
  // Empty when the task is unassigned.
  assigneeId: string;
  // The name Work's task list shows for the assignee. Empty when the task is
  // unassigned or the name is unknown.
  assigneeName: string;
}

// Invoices tagged with the project, in minor units of the insights default
// currency, like the overview's invoice cards.
export interface ProjectInvoiceTotals {
  invoiceCount: number;
  invoiced: number;
  paid: number;
  outstanding: number;
  overdue: number;
  overdueCount: number;
}

export interface ProjectDashboardResponse {
  project: DashboardProjectOption;
  totalTasks: number;
  completedTasks: number;
  progressPct: number;
  overdueTaskCount: number;
  overdueTasks: ProjectOverdueTask[];
  // Minutes logged this month to date, against the same span last month.
  hoursThisMonth: CountCard;
  priorMinutes: number;
  allTimeMinutes: number;
  // Absent when the viewer may not view finance invoices.
  invoices?: ProjectInvoiceTotals;
}

export interface InsightsDashboardResponse {
  revenue: MoneyCard;
  customers: CountCard;
  orders: CountCard;
  margin: PctCard;
}

export interface RevenuePoint {
  date: string;
  amountMinorUnits: number;
}

export interface RevenueSeriesResponse {
  series: RevenuePoint[];
}

export interface AppRevenueShare {
  app: string;
  amountMinorUnits: number;
  pct: number;
}

export interface RevenueByAppResponse {
  byApp: AppRevenueShare[];
}

export interface RevenueComparisonResponse {
  thisPeriod: number;
  lastPeriod: number;
  deltaPct: number;
}

export interface CustomerKPIResponse {
  total: number;
  newThisMonth: number;
  active: number;
  churned: number;
}

export interface TopCustomer {
  customerId: number;
  name: string;
  email: string;
  lifetimeSpendMinorUnits: number;
  orderCount: number;
}

export interface TopCustomersResponse {
  customers: TopCustomer[];
}

export interface LTVBucket {
  label: string;
  lowerMinorUnits: number;
  upperMinorUnits: number;
  customerCount: number;
}

export interface CustomerLTVResponse {
  buckets: LTVBucket[];
}

export interface ChurnWindow {
  windowDays: number;
  churnedCount: number;
  activeCount: number;
  churnPct: number;
}

export interface CustomerChurnResponse {
  windows: ChurnWindow[];
}

export interface ProductKPIResponse {
  totalSkus: number;
  activeSkus: number;
  outOfStock: number;
  avgMarginPct: number;
}

export interface BestSeller {
  productId: number;
  name: string;
  unitsSold: number;
  revenueMinorUnits: number;
}

export interface BestSellersResponse {
  products: BestSeller[];
}

export interface MarginBucket {
  label: string;
  lowerPct: number;
  upperPct: number;
  productCount: number;
}

export interface ProductMarginsResponse {
  buckets: MarginBucket[];
}

export interface SlowMover {
  productId: number;
  name: string;
  stockQuantity: number;
}

export interface SlowMoversResponse {
  products: SlowMover[];
}

export interface CategorySales {
  categoryId: number;
  categoryName: string;
  unitsSold: number;
  revenueMinorUnits: number;
}

export interface ProductCategoryMixResponse {
  categories: CategorySales[];
}

export interface CashFlowPoint {
  date: string;
  inflowMinorUnits: number;
  outflowMinorUnits: number;
}

export interface CashFlowResponse {
  series: CashFlowPoint[];
}

export interface ExpenseByCategory {
  categoryId: number;
  categoryName: string;
  amountMinorUnits: number;
}

export interface ExpensesByCategoryResponse {
  categories: ExpenseByCategory[];
}

export interface InvoiceAgingBucket {
  bucket: string;
  invoiceCount: number;
  amountMinorUnits: number;
}

export interface InvoiceAgingResponse {
  buckets: InvoiceAgingBucket[];
}

export interface OperationsKPIResponse {
  activeStaff: number;
  shiftsThisWeek: number;
  pendingTimeOff: number;
  openLeaveRequests: number;
}

export interface HeadcountResponse {
  activeStaff: number;
  momDelta: number;
  momDeltaPct: number;
}

export interface ShiftMetricsResponse {
  scheduled: number;
  completed: number;
  noShows: number;
  completionRatePct: number;
}

export interface TimeOffResponse {
  pending: number;
  approved: number;
}

export interface UtilizationResponse {
  scheduledHours: number;
  availableHours: number;
  utilizationPct: number;
}

export interface ReportTemplate {
  id: string;
  name: string;
  description: string;
}

export interface ReportTemplatesResponse {
  templates: ReportTemplate[];
}

export interface GenerateReportResult {
  reportId: string;
  url: string | null;
  status: string;
}

export interface InsightsSettingsResponse {
  defaultCurrency: string;
  defaultPeriod: string;
  comparePrevious: boolean;
  dashboardWidgets: string[];
  favoriteReports: string[];
  // insights.setting.v1.InsightsSettings has no fiscal-year-start or
  // timezone field - these are UI-only and never persisted server-side.
  fiscalYearStart: string;
  timezone: string;
}

export interface ScheduledReport {
  id: string;
  templateId: string;
  cadence: string;
  channel: string;
  recipient: string;
  isActive: boolean;
  createdAt: string;
}

export interface ScheduledReportsResponse {
  scheduledReports: ScheduledReport[];
}
